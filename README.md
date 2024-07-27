# rlimit

`rlimit` offers a light TCP server capable of controlling rate-limits of outgoing requests to third party APIs. It uses a token bucket algorithm to control the allowed request rate.


## Usage

1. Clone this repo
2. Build the server `go build -o ./rlimit`
3. Run the server `./rlimit 4000` and the server will listen to tcp connections at port 4000

## Client

1. Start a TCP connection `nc -v localhost 4000`
2. Start a bucket `INIT req01 buc-id 4 4000`:
```
INIT   -> initiates a token bucket
req01  -> request identifier
buc-id -> bucket id
4      -> number of allowed tokens
4000   -> refill rate in miliseconds (every 4000ms a new token will be generated)
```
3. Request tokens for bucket `buc-id` (`CHECK req02 buc-id-1`). The client will receive a `req02: AVAILABLE` when the bucket has remaining tokens

---

Inspired by https://github.com/Mohamed-khattab/Token-bucket-rate-limiter