# go-vhostd

Simple virtual host HTTP request solver.

Edit Configuration
```
"vhosts": [
	{
		"vhost": "example.com",
		"host": "127.0.0.1:5000"
	},
	{
		"vhost": "otherexample.nu",
		"host": "127.0.0.1:8000"
	}
]
```
To start proxying http requests from <vhost> to <host>
