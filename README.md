# gh-proxy-go
A GitHub reverse proxy written in Go

## Building

Initializing the module:

```shell
go mod init gh-proxy-go
```

```shell
go mod tidy
```

## Building

```shell
go build
```

## Running

```she
./gh-proxy-go
```

## Setting up a system service

Copy the executable to a dedicated location

```shell
cp gh-proxy-go /usrr/local/bin
```

Create and edit the service config

```shell
vim /lib/systemd/system/gh-proxy-go.service
```

Content of the config

```
[Unit]
Description=GitHub Reverse Proxy Service
After=network.target

[Service]
Type=simple
User=nobody
Group=nogroup
LimitNOFILE=32768
ExecStart=/usrr/local/bin/gh-proxy-go
AmbientCapabilities=CAP_NET_BIND_SERVICE
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=gh-proxy-go

[Install]
WantedBy=multi-user.target
```

Reload the service daemon

```shell
systemctl daemon-reload
```

Enable the service

```shell
systemctl enable gh-proxy-go
```

Start the service

```shell
systemctl start gh-proxy-go
```

## Setting up Nginx

Sample config in ```/etc/nginx/sites-available/default```

```nginx
server {
    listen 80;
    server_name example.com;
    return 301 https://$server_name$request_uri;  # Redirect HTTP to HTTPS
}
server {
	listen 443 ssl;
	server_name example.com;

	root /var/www/html;

	# Add index.php to the list if you are using PHP
	index index.html index.htm index.nginx-debian.html;

	server_name _;
	ssl_certificate /path/to/cert;
	ssl_certificate_key /path/to/key;
	
	ssl_protocols TLSv1.2 TLSv1.3;
	ssl_prefer_server_ciphers on;
	ssl_ciphers "HIGH:!aNULL:!MD5";
	
	location / {
		# First attempt to serve request as file, then
		# as directory, then fall back to displaying a 404.
		try_files $uri $uri/ =404;
	}

	location /gh/ {
		proxy_pass http://127.0.0.1:8001/;
	}
}
```

## Sending a Request

Now you can try accessing your URL with the GitHub link at the end after the ```/gh``` path.

For example:

```
https://example.com/gh/https://raw.githubusercontent.com/repo/file.txt
```

