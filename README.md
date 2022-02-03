µdensity
========

Do stuff


Paths
-----

```
POST /service/{service}/{projet}/{branch}
    return run id

GET /service/{service}/{projet}/{branch}

GET /service/{service}/{projet}/{run}
GET /service/{service}/{projet}/{run}/_status
GET /service/{service}/{projet}/latest
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
