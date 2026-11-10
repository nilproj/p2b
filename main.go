package main

import (
	"log"
	"os"

	"github.com/nilproj/p2b/generator"
	"github.com/nilproj/p2b/jsonrpc"
)

func main() {
	srv := jsonrpc.NewServer(os.Stdin, os.Stderr)

	srv.Handle("getManifest", handleManifest)
	srv.Handle("generate", handleGenerate)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func handleManifest(w jsonrpc.Writer, r *jsonrpc.Request) {
	w.Write(map[string]any{
		"manifest": map[string]any{
			"prettyName":    "Prisma to Bun",
			"defaultOutput": "../model",
		},
	})
}

func handleGenerate(w jsonrpc.Writer, r *jsonrpc.Request) {
	file, err := os.Create("prisma.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if _, err := file.Write(r.Params); err != nil {
		log.Fatal(err)
	}

	if g, err := generator.New(r.Params); err != nil {
		log.Fatal(err)
	} else if err := g.Generate(); err != nil {
		log.Fatal(err)
	}

	w.Write(nil)
}
