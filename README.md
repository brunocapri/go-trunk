# Go Trunk

Go Trunk is a small, concurrent, in memory cache with sharding.


```
  go get github.com/brunocapri/go-trunk
```

```go 
cache, err := trunk.NewCache[int](time.Second*10, -1)
cache.Add("myKey", 10)
// ...
val, ok := cache.Get("myKey")
```
