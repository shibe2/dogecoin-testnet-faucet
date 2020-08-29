FROM golang AS build
RUN ["mkdir", "-p", "/out/data", "/out/lib/x86_64-linux-gnu", "/out/lib64"]
RUN cp /lib/x86_64-linux-gnu/libdl.so* /lib/x86_64-linux-gnu/libdl-*.so /lib/x86_64-linux-gnu/libpthread.so* /lib/x86_64-linux-gnu/libpthread-*.so /lib/x86_64-linux-gnu/libc.so* /lib/x86_64-linux-gnu/libc-*.so /lib/x86_64-linux-gnu/ld-*.so /out/lib/x86_64-linux-gnu
RUN cp /lib64/ld-linux-x86-64.so* /out/lib64
WORKDIR /backend
COPY go.mod go.sum ./
RUN ["go", "build", "github.com/mattn/go-sqlite3", "gopkg.in/yaml.v3"]
COPY . .
COPY faucetd.yaml /out/
RUN ["go", "build", "-o", "/out/faucetd", "faucet/cmd/faucetd"]

FROM scratch
COPY --from=build /out /
WORKDIR /data
RUN ["/faucetd", "db", "create", "/faucetd.yaml"]
EXPOSE 5000
ENTRYPOINT ["/faucetd", "serve", "/faucetd.yaml"]
