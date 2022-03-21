µdensity
========

Gitlab CI triggers asynchone REST analysis and display [badges](https://github.com/narqo/go-badge) and files.

Paths
-----

```
POST /service/{service}/{projet}/{branch}
    return run id

GET /service/{service}/{projet}/{branch}/{commit}

GET /service/{service}/{projet}/{branch}/latest
GET /service/{service}/{projet}/
``` 

Big Picture
-----------

```
                                                                 +---------+   +--------+
  Gitlab CI                                                  +-->| Compose +-->| Docker |
 +-----------------------------------------------+           |   +-------+-+   +--------+
 |                                               |      +----+-----+     |
 | curl --header 'PRIVATE-TOKEN: ${CI_JOB_JWT}' -+----->| µdensity |     +--> Volumes
 |                                               |      +----------+             |
 +-----------------------------------------------+             HTTP          +---v----+
                                                   <-------------------------| Badges |
                                                                             +--------+
```

Bring your own services
-----------------------

[Services documentation](SERVICES.md)

Manage your service
-------------------

### Admin endpoint

For the Admin http server, use `admin_listen` setting.

* `/` Home page
* `/metrics` Prometheus endpoint
* `/status` Microdensity ping Docker and Gitlab

### Sentry

Sentry is used with zap logging.

Use `SENTRY_DSN` env for setting Sentry errors report.
