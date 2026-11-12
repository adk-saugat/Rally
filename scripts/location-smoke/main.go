// Run against the local development stack and a gateway on port 18080.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	a "github.com/adk-saugat/rally/gen/auth/v1"
	g "github.com/adk-saugat/rally/gen/group/v1"
	l "github.com/adk-saugat/rally/gen/location/v1"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"net"
	"net/http"
	"os"
	"time"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func check(ok bool, label string) {
	if !ok {
		panic("FAIL: " + label)
	}
	fmt.Println("PASS:", label)
}
func conn(port string) *grpc.ClientConn {
	c, e := grpc.NewClient("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	must(e)
	return c
}
func auth(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

type point struct {
	Group string  `json:"group_id"`
	User  string  `json:"user_id"`
	Lat   float64 `json:"latitude"`
	Lon   float64 `json:"longitude"`
	TS    int64   `json:"timestamp_unix_ms"`
}
type peer struct {
	send  func(point)
	recv  func() point
	close func()
}

func main() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
			os.Exit(1)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	ac, gc, lc := conn("50051"), conn("50052"), conn("50053")
	defer ac.Close()
	defer gc.Close()
	defer lc.Close()
	au, gr, lo := a.NewAuthServiceClient(ac), g.NewGroupServiceClient(gc), l.NewLocationServiceClient(lc)
	run := fmt.Sprint(time.Now().UnixNano())
	ids := []string{}
	tokens := []string{}
	for _, name := range []string{"Alice", "Bob", "Charlie"} {
		email := "location-smoke-" + name + "-" + run + "@example.test"
		s, e := au.SignUp(ctx, &a.SignUpRequest{FirstName: "Mock " + name, Email: email, Password: "Mock-Location-123!"})
		must(e)
		login, e := au.Login(ctx, &a.LoginRequest{Email: email, Password: "Mock-Location-123!"})
		must(e)
		ids = append(ids, s.UserId)
		tokens = append(tokens, login.Token)
	}
	group, e := gr.CreateGroup(auth(ctx, tokens[0]), &g.CreateGroupRequest{Name: "Mock trip " + run})
	must(e)
	_, e = gr.JoinGroup(auth(ctx, tokens[1]), &g.JoinGroupRequest{InviteCode: group.InviteCode})
	must(e)
	other, e := gr.CreateGroup(auth(ctx, tokens[2]), &g.CreateGroupRequest{Name: "Isolated mock trip " + run})
	must(e)
	fmt.Println("Fixtures:", run, "group", group.GroupId, "other", other.GroupId, "users", ids)
	r := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer r.Close()
	keys := []string{}
	for _, id := range ids {
		keys = append(keys, "group:"+group.GroupId+":user:"+id)
	}
	keys = append(keys, "group:"+other.GroupId+":user:"+ids[2])
	defer r.Del(context.Background(), keys...)
	ws := func(user int, groupID string) peer {
		h := http.Header{}
		h.Set("Authorization", "Bearer "+tokens[user])
		c, _, e := websocket.DefaultDialer.Dial("ws://localhost:18080/groups/"+groupID+"/ws", h)
		must(e)
		return peer{func(p point) { must(c.WriteJSON(p)) }, func() point {
			must(c.SetReadDeadline(time.Now().Add(5 * time.Second)))
			var p point
			must(c.ReadJSON(&p))
			return p
		}, func() { c.Close() }}
	}
	rpc := func(user int, groupID string) peer {
		sc, stop := context.WithCancel(auth(ctx, tokens[user]))
		s, e := lo.StreamLocation(sc)
		must(e)
		return peer{func(p point) {
			must(s.Send(&l.LocationUpdate{GroupId: groupID, UserId: "forged", Latitude: p.Lat, Longitude: p.Lon, TimestampUnixMs: p.TS}))
		}, func() point {
			p, e := s.Recv()
			must(e)
			return point{p.GroupId, p.UserId, p.Latitude, p.Longitude, p.TimestampUnixMs}
		}, stop}
	}
	seed := func(p peer) { p.send(point{Lat: 41.88, Lon: -87.62, TS: time.Now().UnixMilli()}) }
	waitStored := func(groupID, id string, ts int64) {
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			v, e := r.Get(ctx, "group:"+groupID+":user:"+id).Bytes()
			if e == nil {
				var p point
				if json.Unmarshal(v, &p) == nil && p.TS == ts {
					return
				}
			}
			time.Sleep(20 * time.Millisecond)
		}
		panic("FAIL: Redis latest point")
	}
	for _, mode := range []string{"grpc", "websocket", "mixed"} {
		var pa, pb peer
		if mode == "grpc" {
			pa = rpc(0, group.GroupId)
			pb = rpc(1, group.GroupId)
		} else {
			pa = ws(0, group.GroupId)
			if mode == "mixed" {
				pb = rpc(1, group.GroupId)
			} else {
				pb = ws(1, group.GroupId)
			}
		}
		seed(pb)
		// The persisted first point proves the subscription was established before publishing.
		deadline := time.Now().Add(5 * time.Second)
		for {
			n, e := r.PubSubNumSub(ctx, "group:"+group.GroupId+":locations").Result()
			must(e)
			if n["group:"+group.GroupId+":locations"] >= 1 {
				break
			}
			if time.Now().After(deadline) {
				panic("subscription readiness timeout")
			}
			time.Sleep(20 * time.Millisecond)
		}
		seed(pa)
		pb.recv()
		for i := 0; i < 3; i++ {
			for sender, pair := range []struct{ from, to peer }{{pa, pb}, {pb, pa}} {
				p := point{Lat: 41.881 + float64(i)*.0005 - float64(sender)*.001, Lon: -87.623 + float64(i)*.0005, TS: time.Now().UnixMilli() + int64(i+sender)}
				pair.from.send(p)
				got := pair.to.recv()
				check(got.Group == group.GroupId && got.User == ids[sender] && got.Lat == p.Lat && got.Lon == p.Lon && got.TS == p.TS, fmt.Sprintf("%s movement %d sender %d", mode, i+1, sender))
				waitStored(group.GroupId, ids[sender], p.TS)
			}
		}
		fmt.Println("PASS:", mode, "latest points persisted; authenticated identities override forged IDs")
		pa.close()
		pb.close()
		time.Sleep(100 * time.Millisecond)
	}
	// Reconnect fresh clients; verify boundary coordinates and generated timestamps.
	pa, pb := ws(0, group.GroupId), ws(1, group.GroupId)
	t0 := time.Now().UnixMilli()
	pb.send(point{Lat: 0, Lon: 0, TS: t0})
	waitStored(group.GroupId, ids[1], t0)
	pa.send(point{Lat: 90, Lon: 180})
	got := pb.recv()
	check(got.Lat == 90 && got.Lon == 180 && got.TS >= t0 && got.TS <= time.Now().UnixMilli(), "reconnection, boundary values, and generated timestamp")
	h := http.Header{}
	h.Set("Authorization", "Bearer "+tokens[2])
	isolated, _, e := websocket.DefaultDialer.Dial("ws://localhost:18080/groups/"+other.GroupId+"/ws", h)
	must(e)
	mark := time.Now().UnixMilli()
	must(isolated.WriteJSON(point{Lat: 1, Lon: 1, TS: mark}))
	waitStored(other.GroupId, ids[2], mark)
	pa.send(point{Lat: 0, Lon: 0, TS: mark + 1})
	got = pb.recv()
	check(got.Lat == 0 && got.Lon == 0, "zero coordinates accepted")
	isolated.SetReadDeadline(time.Now().Add(time.Second))
	_, _, readErr := isolated.ReadMessage()
	ne, ok := readErr.(net.Error)
	check(ok && ne.Timeout(), "no cross-group delivery or self-echo during 1 second")
	isolated.Close()
	pa.close()
	pb.close()
	// Fresh WebSocket streams for rejected payloads and membership.
	for _, tc := range []struct {
		name    string
		user    int
		payload any
		code    int
	}{{"nonmember", 2, map[string]any{"latitude": 41.88, "longitude": -87.62}, 1008}, {"invalid latitude", 0, map[string]any{"latitude": 91, "longitude": 0}, 1007}, {"invalid longitude", 0, map[string]any{"latitude": 0, "longitude": 181}, 1007}, {"missing latitude", 0, map[string]any{"longitude": 0}, 1007}} {
		h := http.Header{}
		h.Set("Authorization", "Bearer "+tokens[tc.user])
		c, _, e := websocket.DefaultDialer.Dial("ws://localhost:18080/groups/"+group.GroupId+"/ws", h)
		must(e)
		must(c.WriteJSON(tc.payload))
		c.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, _, e = c.ReadMessage()
		check(websocket.IsCloseError(e, tc.code), tc.name+" rejected")
		c.Close()
	}
	for _, token := range []string{"", "invalid"} {
		h := http.Header{}
		if token != "" {
			h.Set("Authorization", "Bearer "+token)
		}
		c, response, e := websocket.DefaultDialer.Dial("ws://localhost:18080/groups/"+group.GroupId+"/ws", h)
		if c != nil {
			c.Close()
		}
		check(e != nil && response != nil && response.StatusCode == 401, "unauthorized handshake rejected")
	}
	n, e := r.Exists(ctx, "group:"+group.GroupId+":user:"+ids[2]).Result()
	must(e)
	check(n == 0, "nonmember position not persisted")
	fmt.Println("PASS: core location sharing works with mock movement over gRPC, WebSocket, and mixed clients")
	fmt.Println("Fixture accounts/groups remain: APIs do not expose deletion; test Redis positions cleaned up.")
}
