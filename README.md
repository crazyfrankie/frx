# frx
This project provides some Go-related encapsulations including:

- Encapsulation of non-business logic code for large-scale project development
  - HTTP middleware:
    - `ctxcache`: provide ctx cache on a single HTTP request link
  - Error Code:
    - `errorx`: provide a complete error code design, supporting registration and dynamic creation
  - log:
    - `logs`: provide a complete logging function 
- Wrappers for frameworks and standard libraries
  - httpx:
    - a wrapping for `http.Client` and `http.Server`
- Third-party toolkits (e.g. `idgen` as a distributed unique ID generator)