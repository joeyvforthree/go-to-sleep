.PHONY: gate vet

vet:
	go vet ./...

# gate is what the global pre-push hook runs before every push. Check-only:
# it must never rewrite a tracked file, or the push is blocked. The live-API
# tests sit behind the integration build tag and are not run here.
gate: vet
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	go test ./...
	go build ./...
