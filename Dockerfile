FROM golang:1.27-alpine AS builder                                                                               
WORKDIR /src                                                                                                     
COPY go.mod go.sum ./                                                                                            
RUN go mod download                                                                                              
COPY . .                                                                                                         
RUN CGO_ENABLED=0 go build -trimpath -o /out/server .                                                            

FROM alpine:3.20                                                                                                 
WORKDIR /app                                                                                                     
COPY --from=builder /out/server /usr/local/bin/server                                                            
COPY config.json /app/config.json                                                                                
EXPOSE 8000                                                                                                      
ENTRYPOINT ["server"]      