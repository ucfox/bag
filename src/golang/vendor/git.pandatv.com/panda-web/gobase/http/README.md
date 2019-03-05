### Base Usage
1. 返回Response
```
httpclient.Get(url)
httpclient.PostForm(url, params)
httpclient.Post(url, bodyType, body)
httpclient.Do(request)
```

2. 返回String
```
httpclient.GetAsString(string)
httpclient.PostFormAsString(url, params)
httpclient.PostAsString(url, bodyType, body)
httpclient.DoAsString(request)
```

3. 返回Json
```
httpclient.GetAsJson(string, interface)
httpclient.PostFormAsJson(url, params, interface)
httpclient.PostAsJson(url, bodyType, body, interface)
httpclient.DoAsJson(request, interface)
```
### 自定义Cliet
默认httpclient请求超时1秒，通过自定义client,可以控制http各参数
```
c := &http.Client{
    Timeout: time.Second*5,
}
cli := httpclient.NewClient(c)
```
