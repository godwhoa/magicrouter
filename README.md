# Magic Router
> Like cloudflare but for OpenAI API. Fallback, rate limiting, logging and more.


# Capability
- [x] Proxying to OpenAI API
- [x] Fallback for providers/models
- [ ] Improve fallback with redis based circuit breaker
- [ ] Azure provider
- [ ] Response logging
- [ ] Rate limiting

# Tech debt
- [x] Test fallback
- [x] Ensure providers return ErrRateLimited etc.
