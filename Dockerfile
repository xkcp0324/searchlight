FROM appscode/icinga:8.0.0-k8s
MAINTAINER xuekui <kui.xue@dmall.com>

COPY bin/hyperalert /usr/lib/monitoring-plugins/
