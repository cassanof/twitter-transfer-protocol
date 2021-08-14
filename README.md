# A basic PoC and library for transporting information using twitter direct messages
### FOR EDUCATIONAL OR RESEARCH USE ONLY
#### Rate Limiting:
The Twitter API is implements to rate limiting. After a certain amount of requests the API sends back error codes (code `88` or `226`).
The only way to circumvent this is either to rotate accounts or to wait until the block gets lifted.  
More info on rate limits: https://developer.twitter.com/en/docs/twitter-api/v1/rate-limits
