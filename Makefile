.PHONY: dist

dev:
	air -build.exclude_dir "dist" -build.cmd "make dist"

dist:
	rm -rf ./dist
	cd ./cmd && GOOS=js GOARCH=wasm go build -o ../bin/quadtree-demo.wasm
	mkdir -p dist/lib
	mkdir -p dist/bin
	cp ./lib/wasm_exec.js dist/lib/wasm_exec.js
	cp ./bin/quadtree-demo.wasm dist/bin/quadtree-demo.wasm
	cp ./index.html dist/index.html

clean:
	rm -rf ./dist
	# rm -rf ./.vercel/output

deploy-dev:
	make clean
	vercel build
	vercel deploy --prebuilt

deploy:
	make clean
	vercel build --prod
	vercel --prebuilt --prod