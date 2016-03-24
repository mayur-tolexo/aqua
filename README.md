# AQUA
Golang Restful APIs in a cup, and ready to serve!

## Inspiration
- Go-Rest framework for service controllers and tag-based configurations
- Popular WebServers (like Apache,IIS) for hierarchical configuration model

## Design Goals

- Simplicity & Modularity
- Developer Productivity
- High Configurability
- Low learning curve
- Easy Versioning
- Pluggable Modules (using golang middleware)
- High Performance
- Preference for Json (over xml)

## Features

- Service Controllers to define endpoints (modular and organized code)
- Powerful Configuration Model. Supports 4 levels:
  1. at server level, programmatically (these are inherited by all endpoints)
  2. at service controller level, declaratively using golang tags (these are inherited by all contained apis)
  3. at service controller level, programmatically
  4. at api or endpoint level, declaratively (these override inherited configurations)
- Versioning
  - Multiple versions are supported easily by
     - defining at service controller level (inherited by all internal endpoints)
     - overriding for each endpoint specifically
- Caching
- Database Binding
 -  CRUD endpoints
 -  Limited ad-hoc querying
- Stubbing
 - If there are code/project dependencies on your api service, you can simply write a stub (sample output) in an external file and publish this mock api quickly before writing actual business logic
- Working with Queues *|pending*
- Proxying or Wrapping around existing APIs *|pending*
- Middleware support (by defining modules)
- Logging (using middleware)

##Lets explore these features

#### Q: How do I write a 'hello world' api?
First define a service controller in your project that supports a GET response (aqua.GetApi as its type). Note that the controller defined as a struct must anonymously include aqua.RestService.

```
type HelloService struct {
	aqua.RestService
	world aqua.GET
}
```

Now implement a method corresponding to 'world' field after uppercasing the first letter. To start off, the method can return a string (more on this later).

```
func (me *HelloService) World() string {
	return "Hello World"
}
```

Now setup your main function to run the Aqua rest server

```
server := aqua.NewRestServer()
server.AddService(&HelloService{})
server.Run()
```

Now open your browser window, and hit http://localhost:8090/hello/world

---


#### Q: But I don't need any magic; What about the unadulterated http requests and responses?

Sure, just change the function signature and you are good to go.

```
func (me *HelloService) World(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello There!")
}
```
---

####Q: I want to change the url from /hello/world to /hello/moon. Do I need to change the method names?

The service urls are derived from url tags. If none are specified then it defaults to the method name. So you can simply introduce the tag as follows.

```
type HelloService struct {
	aqua.RestService
	world aqua.GET `url:"moon"`
}
```

---

#### Q: What if I need to return both Hello World, and Hello There as different versions of the same GET api?

Simply add both the methods, but specify versions in field tags.

```
type HelloService struct {
	aqua.RestService
	world aqua.GET `version:"1.0" url:"moon"`
	worldNew aqua.GET `version:"1.1" url:"moon"`
}
func (me *HelloService) World() string {
	return "Hello World"
}
func (me *HelloService) WorldNew(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello There!")
}
```
Now you can hit:

http://localhost:8090/v1.0/hello/moon and

http://localhost:8090/v1.1/hello/moon to see the difference.

---

#### Q: What options can I use to customize URLs for my apis?

There are 3 out-of-box setting available, that help you customize URLs.

- prefix
- root
- url

We have already seen how 'url' works.

To change the root directory (*hello*), you can use the *root* tag at each service level, or more simply at the service controller level as demonstrated below:

```
type HelloService struct {
	aqua.RestService  `root:"this-is-the"`
	world aqua.GET `version:"1.0" url:"moon"`
	worldNew aqua.GET `version:"1.1" url:"moon"`
}
```

With this change, your api endpoints are now working as:

*http://localhost:8090/v1.0/this-is-the/moon* and

*http://localhost:8090/v1.1/this-is-the/moon*

You can also use the 'prefix' field. This part comes in before version information in the final constructed endpoint url

```
type HelloService struct {
	aqua.RestService  `root:"this-is-the" prefix:"sunshine"`
	world aqua.GET `version:"1.0" url:"moon"`
	worldNew aqua.GET `version:"1.1" url:"moon"`
}
```
So with this prefix now set, our end points would become:

*http://localhost:8090/sunshine/v1.0/this-is-the/moon*

*http://localhost:8090/sunshine/v1.1/this-is-the/moon*

Also note that, all there of these properties (url, root and prefix) can contain any number of slashes. So if you change the url to:

```
type HelloService struct {
	aqua.RestService  `root:"this-is-the" prefix:"sunshine"`
	world aqua.GET `version:"1.0" url:"/good/old/moon"`
}
```

Then you get the final url as:

http://localhost:8090/sushine/v1.0/this-is-the/good/old/moon.

---

#### Q: Does Aqua use any mux?

Yes, Gorilla mux is used internally. So to define url parameters, we'll need to follow Gorilla mux conventions. We'll get to those in a moment

---

#### Q: How can I check if the server is up and running?

By default an "aqua" route is setup:

 - */aqua/ping* returns "pong" if the server is running
 - */aqua/status* returns version, go runtime memory information
 - */aqua/time* returns current server time

#### Q: What is the default port that Aqua runs on?

It's 8090. You can change it though as follows:

```
server := aqua.NewRestServer()
server.AddService(&HelloService{})
server.Port = 5432;
server.Run()
```

---

#### Q: When I use api versioning, can I use HTTP headers to pass the version info?

```
type CatalogService struct {
	aqua.RestService  `root:"catalog" prefix:"mycompany"`
	getProduct aqua.GET `version:"1.0" url:"product"`
}
```

If you setup a catalog service as shown above then out of box you can use version capability as shown below

1. GET call to http://localhost:8090/mycompany/v1.0/catalog/product
2. GET call to http://localhost:8090/mycompany/catalog/product
  - pass a request header "Accept": "application/vnd.api+json;version=1.0"
  -  *-or-*
  - pass a request header "Accept": "application/vnd.api-v1.0+json"

Note: If you want to customize the media type, you can do so.

```
type CatalogService struct {
	aqua.RestService  `root:"catalog" prefix:"mycompany"`
	getProduct aqua.GetApi `vendor:"vnd.myorg.myfunc.api" version:"1.0" url:"product"`
}
```
Basis this, the required Accept header will be need to changed to following:

- "Accept" header : "application/__vnd.myorg.myfunc.api__+json;version=1.0"
-  *-or-*
- "Accept" header : "application/__vnd.myorg.myfunc.api__-v1.0+json"

---

#### Q: How can I access query strings?

Its simple, you add an input variable to your implementation method of type aqua.Aide (a helper class) This variable gives you access to the Request object, and also has some helper methods as shown below:

```
type HelloService struct {
	aqua.RestService
	world aqua.GET
}

func (me *HelloService) World(j aqua.Aide) string {
	j.LoadVars()
	return "Hello " + j.QueryVars["country"]
}
```

Now, just hit the url: http://localhost:8090/hello/world?country=Singapore

---

#### Q: How do I pass dynamic parameters to apis?

You start by defining the url with the appropriate dynamic variable as per the guidelines of Gorilla mux.

```
type HelloService struct {
	aqua.RestService
	world aqua.GET `url:"/country/{c}"`
}
```
Then you just read this value in the associated method. Note: Aqua currently supports passing int and string parameters.

```
func (me *HelloService) World(c string) string {
	return "Hello " + c
}
```

Now, you can hit http://localhost:8090/hello/country/Brazil

In case you are reading an integer value, then you can define strict logic in url to only match numbers using a regular expression:

```
type HelloService struct {
	aqua.RestService
	world aqua.GET `url:"/country/{c}"`
	capital aqua.GET `url:/capital/{cap:[0-9]+}`
}
```
---

#### Q: Can you explain how the configuration model works? Will I need to define attributes at each endpoint level?

Aqua has a powerful configuration model that works at 4 levels:

1. Server (programmatically)
2. Service controller (declaratively)
3. Service controller (programmatically)
4. Endpoint (declaratively)

Lets look at each of them in detail

---

###### 1. Server (programmatically)

If you define any configuration at the server level, then it is __inherited__ by all the Service controllers and all the contained services automatically.

```
server := aqua.NewRestServer()

// Note:
server.Prefix = "myapis"
// Prefix value is inherited by everything on this server!!

server.AddService(&HelloService{})
server.AddService(&HolaService{})
server.Run()
```
---

###### 2. Service controller (declaratively)

We added two service controllers to the server above - HelloService and HolaService. Let's assume that all the contained services need to begin with words 'Hello' and 'Hola' respectively.

To achive this, we specify the 'root' variable at the top level by defining it agains the RestServer.

```
type HelloService struct {
	aqua.RestService `root:"Hello"`
	service1 aqua.GET
	service2 aqua.GET
	..
	serviceN aqua.GET
}
```

This ensures that all services in this now __inherit__ the root value of "Hello"

---

###### 4. Endpoint (declaratively)

Last but not the least, you can specify a value at a service endpoint. You can do so by configuring at the api level as shown below. Note that these values will override the inherited values.

```
type HelloService struct {
	aqua.RestService `root:"Hello"`
	service1 aqua.GET `root:"Hiya"` //Hiya overrides Hello
	service2 aqua.GET
	..
	serviceN aqua.GET
}
```

---

#### Q: What all configurations are available in Aqua?

| Tag          | Usage            
| ----------   |-----------------
| prefix, pre  | Url prefix as in: http://abc.com/[prefix]/v1/root/url                 
| root         | Url root as in: http://abc.com/prefix/v1/[root]/url                                  
| url          | Url path as in in: http://abc.com/prefix/v1/root/[url]                            
| version, ver | Url version as in: http://abc.com/prefix/v[1]/root/url                                  
| vendor, vnd  |
| modules, mods| Sequence of module names (or middlewares) that the request goes through                 
| cache        | The name of cache provider to use
| ttl          | Duration to cache (e.g. 5s or 10m)
| stub         | Relative or absolute path to the file containing the mock stub
| wrap         | Wrapping other/3rd party rest services

---

#### Q: Is caching supported? How do I configure it?

Multiple cache provider's can be added to RestServer via its AddCache method. The arguments are the unique name of the cache and an implementation of Cacher interface. Note: this interface is defined in the sister library aero/cache

```
server := aqua.NewRestServer()
server.AddCache("mycache", <implementation of aero.cache.Cacher>)
server.Run()
```

This interface is defined as:

```
// package aero.cache
type Cacher interface {
	Set(key string, data []byte, expireIn time.Duration)
	Get(key string) ([]byte, error)
}
```

"aero.cache" also defined multiple implementations of this interface, namely:

1. Memcache implementation
2. Redis implementation
3. In memory cache (recommended for dev boxes)
4. Debug Wrapper to log reads and writes to text files (for debugging)

Let's utilize the memcache implementation.

```
server := aqua.NewRestServer()
server.AddCache("mycache", cache.NewMemcache("127.0.0.1", 11211))
server.AddCache("remote", cache.NewMemcache("123.234.012.23", 11211))
server.Cache = "mycache" // default for all endpoints
server.Run()
```

Now all we need to do to use this cache is to set the "ttl" tag. So lets look at some services

```
type CatalogService struct {
	RestService
	getProduct  GET `url:"product/{id}"  ttl:"5m"`
	getSeller   GET `url:"seller/{id}"   ttl:"15m" cache:"remote"`
}
```
Thats it. You are good to go!

Service http://localhost:8090/catalog/product/{id} inherts cache as "mycache" from server. All calls to different invocations will get cached for 5min (as set in ttl).

Service http://localhost:8090/catalog/seller/{id} overrides the server cache and sets its cache store to be "remote". So, all invocations will be saved in this remote memcache instance for a rolling 15min duration.

Current limitations:

- only available for GET calls (others will be added)
- not supported when using standard http handler (w http.ResponseWriter, r *http.Request)

Note: the cache key is the the unique url of the request (including the query string parameters)

---

#### Q: What are 'modules' and how can I use them? (Or, does Aqua support Golang middleware?)

Modules allow you to harness the power of Golang middleware. A module is a function that returns an anonymous function that consumes an http.Handler and returns another http.Handler. The signature of the module would be:

```
func SomeFunc(<params_if_needed>) func(http.Handler) http.Handler {
  // code
}
```
Say, we want to log all those calls that are taking more then 1 sec to return. So we write this module as:

```
func LogSlowCalls() func(http.Handler) http.Handler {
  file := "/path"
  f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
  if err != nil {
     // error handling..
  }
  l := log.New(f, "", log.LstdFlags)
  return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			dur := time.Since(start).Seconds()
			if dur >= 1 {
				l.Printf("%s %s %.3f", r.Method, r.RequestURI, time.Since(start).Seconds())
			}
		})
	}
}
```
We can also pass duration to log and file path as parameters to LogSlowCalls function to make it more customizable. Now lets add this module to our server.

```
server := aqua.NewRestServer()
server.AddModule("slowLog", LogSlowCalls("/tmp/file.log", 1))
server.Run()

// And to use it in our service, we just pass it to the tag
type CatalogService struct {
	RestService
	getProduct  GET `url:"product/{id}"  module:"slowLog"`
	getSeller   GET `url:"seller/{id}"`  module:"slowLog, module2, module3"
}

```

Note:

- slowLog is set for getProduct endpoint. Any request that takes more than 1 sec will be logged.
- getSeller endpoint has multiple modules set in a comma seprated manner. All these will be invoked in the sequence slowLog (first) → module2 → module3 (last) → and finally the method "GetProduct".
- Aqua uses **github.com/carbocation/interpose** for method chaining internally to setup these golang middleware.

---

#### Q: Are there any out-of-box modules bundled with Aqua?

Just a few:

- A slow logger, ModSlowLog, that takes millisec precision as an input
- An access logger, ModAccessLog

---

#### Q: Can I create mock apis to de-bottleneck development?

Yes. Aqua makes is possible to create mock api stubs using external files. You can specify an associated file using the "stub" tag as shown below.

```
type MockService struct {
	RestService
	yetToCode  GET `stub:"samples/some.json"`
}

// And then run it
server := aqua.NewRestServer()
server.AddService(&MockService{})
server.Run()
```
Now when you invoke http://localhost:8090/mock/yet-to-code then the contents of "samples/some.json" file are read and returned. Note that the returned code may not necessarily be json data. It can be anything.

Aqua searches for file in both:

- executable directory, and
- working directory


Also, you can specify file path using both:

- relative file syntax, and
- absolute file path syntax

If the file is not found then Aqua returns 400 status code.

When using such mock stubs, you don't need to define any methods for your endpoints.

---

#### Q: Is there any support for creating CRUD api's out of the box?

In order to improve developer productivity Aqua supports basic database operations. This functionality is tied to the popular GORM https://github.com/jinzhu/gorm project. Let us see how we can set this up, for a "user" table

First, we define the model (as per GORM specs):

```
type User struct {
	id       int `gorm:"primary_key"`
	username string
	name     string
}
```

Second, we add an endpoint of type CrudApi (instead of Post or Get etc)

```
type AutoService struct {
	RestService
	users  CRUD
}
```

Third, in the service method, we define the function to return a CrudApi struct address.

```
func (s *AutoService) Users() CRUD {
	return CRUD {
		Model: func() (interface{}, interface{}) {
			return &User{}, nil
		}
	}
}
```

Basically in the CrudApi struct we define the gorm model to use as an address to return in the Model() function.

Now let's test the ready-made endpoints, by hitting:

- GET to http://localhost:8090/auto/users/123
- POST to http://localhost:8090/auto/users such that they body payload contains a json like:

```
{
	id: 1234,
	username: "jdoe",
	name: "John Doe"
}
```

- PUT to http://localhost:8090/auto/users/345 such that the body payload contains a json to update user with id 345

```
{
	username: "jbrown",
	name: "Jason Browne"
}
```
- DELETE to http://localhost:8090/auto/users/567

That's it. You write a function to return a CrudApi object and you get 4 CRUD methods out of the box.

By default, AQUA uses the default master database (as specified in you yaml file). If you want to override it then you can do so easily, as shown below:

```
func (s *AutoService) Users() CRUD {
	return CRUD {
		Engine: "mysql",
		Conn: "your-connection-string",
		Model: func() (interface{}, interface{}) {
			return &User{}, nil
		}
	}
}

```

---

#### Q: CRUD is nice but also limiting. Does Aqua support ad-hoc querying?

While fully generic ad-hoc querying (that may include joins over many tables) is not yet supported, AQUA does support limited querying to a specific model.

In other words if your query is of this type, then you can run it out of box with AQUA.

```
SELECT * FROM <model_table> WHERE <conditions>
```

Continuing from the previous example, let us modify the CrudApi return to include Models() method:

```
func (s *AutoService) Users() CRUD {
	return CRUD {
		// Add 2nd return (slice of your models)
		Model: func() (interface{}, interface{}) {
			return &User{}, &[]User{}
		}
	}
}
```
This will give you two additional endpoints:

- POST @ http://localhost/auto/users/! 
- POST @ http://localhost/auto/users/$

Let us see each of these in detail.

__POST @ http://localhost/auto/users/!__

This endpoint takes SQL where cluase as Raw Body and returns a json array of matching users. So, we can the Body as:

```
id in (1,2,3,4,5) OR username like "j%"
```

When executed, the final query becomes:

```
SELECT * FROM users WHERE id in (1,2,3,4,5) OR username like "j%"
```

__POST @ http://localhost/auto/users/$__

This endpoint takes parameterized inputs. You specify a json as Raw Body with "where" and "params" keys as shown below:

```
	{
		"where"  : "username like ? or name like ?",
		"params" : [ "j%", "Tim%" ],
		"limit"  : 100,
		"offset" : 25,
		"order"  " ["username", "name desc"]
	}
```

When executed, the final query becomes:

```
SELECT * FROM users WHERE username like "j%" or name like "Tim%" order by username, name desc limit 100 offset 25
```

The output is same, a json array.

---

#### Q: CRUD works for RDBMS only or supports NoSQL systems?


Talking to NoSQL systems is planned but not implemented yet.

---


#### Q: If I wanted to switch out gorm (default ORMapping tool used), and switch to a different one then can that be achieved?

Yes, this can be done. At this time however, only GORM is supported.

---


#### Q: Is there any support to wrap or proxy to other/3rd party api calls using Aqua?

There are a number of good reasons why you would want to wrap any existing Rest apis, say:

- to enable caching
- to setup logging / monitoring
- to modify/manipulate responses or headers

Aqua supports this using the "wrap" tag configuration

```
type WrapperService struct {
	RestService
	MyDoSomething  GetApi `wrap:"http://abc.com/do/someting" ttl:"25m"`
}

```
[TBD]


---

#### Q: Can I manipulate the response headers or response body?

You can add a header through a middleware. But this must be done before the call to next.ServeHTTP(..). Any header additions or response body modifications post this will have no effect.

```
func ModModify() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom-A", "a1") /* this works */
			next.ServeHTTP(w, r)
			w.Header().Set("X-Custom-B", "b1") /* no effect */
		})
	}
}
```


To conditionally add headers or modify response basis the body, you need to use httptest.ResponseRecorder. There is already a built in middleware that gives this functionality - it is named ModRecorder(). Let us see how to use it, with a complete example.



```
func ModModify() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom-A", "a1")
			next.ServeHTTP(w, r)
			w.Header().Set("X-Custom-B", "b1")
			w.Write([]byte("appending text to response body"))
		})
	}
}
```
Now we add this to our server along with ModRecorder:

```
	s := aqua.NewRestServer()
	s.AddModule("chg", ModModify())
	s.AddModule("rec", aqua.ModRecorder())
	s....
	s.Run()
```

Now invoke both these middleware in your rest call.

```
	type HelloService struct {
	aqua.RestService
	world  aqua.GetApi `mods:"rec,chg"`
}
```
This will first call ModRecorder middleware which will setup a httptest.ResponseRecorder and pass it on. The next middleware, ModModify can now change headers before and after the next.ServeHTTP call. It is also able to modify the response body now.

---


#### Q: Lot of the examples on this page use "string" as the return type for method implementations. Are there other data types allowed?

---

#### Q: I would like to run some cron jobs. I could write a separate application and run it through crontab. Or I could invoke a separate Aqua service through a "curl" call.


