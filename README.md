# AQUA
Golang Restful APIs in a cup, and ready to serve!

## Features
- Versioning
  - Multiple versions are supported easily by
     - defining at service controller level (inherited by all internal endpoints)
     - overriding for each endpoint specifically
- Database Binding
 -  CRUD endpoints
 -  Limited ad-hoc querying
- Working with Queues *|pending*
- Proxying or Wrapping around existing APIs *|pending*
- Middleware support (by defining modules)
- Logging (using middleware)

##Lets explore these features

#### Q: How can I check if the server is up and running?

By default an "aqua" route is setup:

 - */aqua/ping* returns "pong" if the server is running
 - */aqua/status* returns version, go runtime memory information
 - */aqua/time* returns current server time


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

#### Q: Are there any out-of-box modules bundled with Aqua?

Just a few:

- A slow logger, ModSlowLog, that takes millisec precision as an input
- An access logger, ModAccessLog

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


