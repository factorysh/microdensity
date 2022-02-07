µdensity
========

Gitlab CI triggers asynchone analysis and display badges and files.

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
