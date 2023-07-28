package main

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	_ "github.com/flonja/multiversion/protocols" // VERY IMPORTANT
	v419 "github.com/flonja/multiversion/protocols/v419"
	v589 "github.com/flonja/multiversion/protocols/v589"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
)

// The following program implements a proxy that forwards players from one local address to a remote address.
func runProxy(config config) {
	var src oauth2.TokenSource
	if config.AuthEnabled {
		src = tokenSource()
	}

	p, err := minecraft.NewForeignStatusProvider(config.Connection.RemoteAddress)
	if err != nil {
		panic(err)
	}
	listener, err := minecraft.ListenConfig{
		StatusProvider:         p,
		AcceptedProtocols:      []minecraft.Protocol{v419.New(), v589.New()},
		AuthenticationDisabled: !config.AuthEnabled,
	}.Listen("raknet", config.Connection.LocalAddress)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	for {
		c, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleConn(c.(*minecraft.Conn), listener, config, src)
	}
}

// handleConn handles a new incoming minecraft.Conn from the minecraft.Listener passed.
func handleConn(conn *minecraft.Conn, listener *minecraft.Listener, config config, src oauth2.TokenSource) {
	serverConn, err := minecraft.Dialer{
		KeepXBLIdentityData: true,
		IdentityData:        conn.IdentityData(),
		ClientData:          conn.ClientData(),
		TokenSource:         src,
	}.Dial("raknet", config.Connection.RemoteAddress)
	if err != nil {
		panic(err)
	}
	var g sync.WaitGroup
	g.Add(2)
	go func() {
		if err := conn.StartGame(serverConn.GameData()); err != nil {
			panic(err)
		}
		g.Done()
	}()
	go func() {
		if err := serverConn.DoSpawn(); err != nil {
			panic(err)
		}
		g.Done()
	}()
	g.Wait()

	go func() {
		defer listener.Disconnect(conn, "connection lost")
		defer serverConn.Close()
		for {
			pk, err := conn.ReadPacket()
			if err != nil {
				return
			}
			if err := serverConn.WritePacket(pk); err != nil {
				if disconnect, ok := errors.Unwrap(err).(minecraft.DisconnectError); ok {
					_ = listener.Disconnect(conn, disconnect.Error())
				}
				return
			}
		}
	}()
	go func() {
		defer serverConn.Close()
		defer listener.Disconnect(conn, "connection lost")

		// r := world.Overworld.Range()
		// biomeBufferCache := make(map[protocol.ChunkPos][]byte)

		for {
			pk, err := serverConn.ReadPacket()

			// switch pk := pk.(type) {
			// case *packet.SubChunk:
			// 	chunkBuf := bytes.NewBuffer(nil)
			// 	blockEntities := make([]map[string]any, 0)
			// 	for _, entry := range pk.SubChunkEntries {
			// 		if entry.Result != protocol.SubChunkResultSuccess {
			// 			chunkBuf.Write([]byte{
			// 				chunk.SubChunkVersion,
			// 				0, // The client will treat this as all air.
			// 				uint8(entry.Offset[1]),
			// 			})
			// 			continue
			// 		}

			// 		var ind uint8
			// 		readBuf := bytes.NewBuffer(entry.RawPayload)
			// 		sub, err := chunk.DecodeSubChunk(latest.NewBlockMapping().Air(), r, readBuf, &ind, chunk.NetworkEncoding)
			// 		if err != nil {
			// 			fmt.Println(err)
			// 			continue
			// 		}

			// 		var blockEntity map[string]any
			// 		dec := nbt.NewDecoderWithEncoding(readBuf, nbt.NetworkLittleEndian)
			// 		for {
			// 			if err := dec.Decode(&blockEntity); err != nil {
			// 				break
			// 			}
			// 			blockEntities = append(blockEntities, blockEntity)
			// 		}

			// 		chunkBuf.Write(chunk.EncodeSubChunk(sub, chunk.NetworkEncoding, chunk.SubChunkVersion9, r, int(ind)))
			// 	}

			// 	chunkPos := protocol.ChunkPos{pk.Position.X(), pk.Position.Z()}
			// 	_, _ = chunkBuf.Write(append(biomeBufferCache[chunkPos], 0))
			// 	delete(biomeBufferCache, chunkPos)

			// 	enc := nbt.NewEncoderWithEncoding(chunkBuf, nbt.NetworkLittleEndian)
			// 	for _, b := range blockEntities {
			// 		_ = enc.Encode(b)
			// 	}

			// 	_ = conn.WritePacket(&packet.LevelChunk{
			// 		Position:      chunkPos,
			// 		SubChunkCount: uint32(len(pk.SubChunkEntries)),
			// 		RawPayload:    append([]byte(nil), chunkBuf.Bytes()...),
			// 	})
			// 	_ = conn.Flush()
			// 	continue
			// case *packet.LevelChunk:
			// 	if pk.SubChunkCount != protocol.SubChunkRequestModeLimitless && pk.SubChunkCount != protocol.SubChunkRequestModeLimited {
			// 		// No changes to be made here.
			// 		break
			// 	}

			// 	max := r.Height() >> 4
			// 	if pk.SubChunkCount == protocol.SubChunkRequestModeLimited {
			// 		max = int(pk.HighestSubChunk)
			// 	}

			// 	offsets := make([]protocol.SubChunkOffset, 0, max)
			// 	for i := 0; i < max; i++ {
			// 		offsets = append(offsets, protocol.SubChunkOffset{0, int8(i + (r[0] >> 4)), 0})
			// 	}

			// 	biomeBufferCache[pk.Position] = pk.RawPayload[:len(pk.RawPayload)-1]
			// 	_ = serverConn.WritePacket(&packet.SubChunkRequest{
			// 		Position: protocol.SubChunkPos{pk.Position.X(), 0, pk.Position.Z()},
			// 		Offsets:  offsets,
			// 	})
			// 	_ = serverConn.Flush()
			// 	continue
			// 	// case *packet.Transfer:
			// 	// 	a.remoteAddress = fmt.Sprintf("%s:%d", pk.Address, pk.Port)

			// 	// 	pk.Address = "127.0.0.1"
			// 	// 	pk.Port = a.localPort
			//}
			if err := conn.WritePacket(pk); err != nil {
				return
			}
			_ = conn.Flush()

			if err != nil {
				if disconnect, ok := errors.Unwrap(err).(minecraft.DisconnectError); ok {
					_ = listener.Disconnect(conn, disconnect.Error())
				}
				return
			}
			if err := conn.WritePacket(pk); err != nil {
				return
			}
		}
	}()
}

type config struct {
	Connection struct {
		LocalAddress  string
		RemoteAddress string
	}
	AuthEnabled bool
}

// tokenSource returns a token source for using with a gophertunnel client. It either reads it from the
// token.tok file if cached or requests logging in with a device code.
func tokenSource() oauth2.TokenSource {
	check := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	token := new(oauth2.Token)
	tokenData, err := os.ReadFile("token.tok")
	if err == nil {
		_ = json.Unmarshal(tokenData, token)
	} else {
		token, err = auth.RequestLiveToken()
		check(err)
	}
	src := auth.RefreshTokenSource(token)
	_, err = src.Token()
	if err != nil {
		// The cached refresh token expired and can no longer be used to obtain a new token. We require the
		// user to log in again and use that token instead.
		token, err = auth.RequestLiveToken()
		check(err)
		src = auth.RefreshTokenSource(token)
	}
	tok, _ := src.Token()
	b, _ := json.Marshal(tok)
	_ = os.WriteFile("token.tok", b, 0644)
	return src
}
