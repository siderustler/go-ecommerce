# run templ generation in watch mode to detect all .templ files and 
# re-create _templ.txt files on change, then send reload event to browser. 
# Default url: http://localhost:7331
live/templ:
	templ generate --watch --proxy="http://localhost:8080" --open-browser=false -v

# run air to detect any go file changes to re-build and re-run the server.
live/server:
	go run github.com/cosmtrek/air@v1.51.0 \
	--build.cmd "go build -o tmp/bin/main" --build.bin "tmp/bin/main" --build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true

# run tailwindcss to generate the styles.css bundle in watch mode.
live/tailwind:
	npx --yes tailwindcss -i ./ports/views/public/input.css -o ./ports/views/public/styles.css --minify --watch=always

# run esbuild to generate the index.js bundle in watch mode.
live/esbuild:
	npx --yes esbuild js/index.ts --bundle --outdir=assets/ --watch

# watch for any js or css change in the assets/ folder, then reload the browser via templ proxy.
live/sync_assets:
	go run github.com/cosmtrek/air@v1.51.0 \
	--build.cmd "templ generate --notify-proxy" \
	--build.bin "true" \
	--build.delay "100" \
	--build.exclude_dir "" \
	--build.include_dir "assets" \
	--build.include_ext "js,css"

# start all 5 watch processes in parallel.
live: 
	make -j4 live/templ live/server live/tailwind live/sync_assets

db:
	docker run -it \
	-e POSTGRES_USER=user \
	-e POSTGRES_PASSWORD=secret \
	-e POSTGRES_DB=ecomm \
	-v $(CURDIR)/sql/01_schema.sql:/docker-entrypoint-initdb.d/01_schema.sql:ro \
	-v $(CURDIR)/sql/02_seed.sql:/docker-entrypoint-initdb.d/02_seed.sql:ro \
	-p 5432:5432 postgres


#need to be authorized to stripe (stripe cli = stripe login)
payment_provider:
	stripe listen --forward-to localhost:8080/api/stripe/wh
