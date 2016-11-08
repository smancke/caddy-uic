caddy-uic
==========

This is a plugin to enable the ui composition features of https://github.com/tarent/lib-compose 
for the [https://caddyserver.com](caddyserver).

Example Caddyfile

```
 uic / http://upstream-host/ {
    fetch layout.html
    fetch http://example.org/navigation
 }
```
