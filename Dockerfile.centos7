FROM openshift/origin-release:golang-1.10 AS build
COPY . /go/src/github.com/openshift/eventrouter
RUN cd /go/src/github.com/openshift/eventrouter && go build .
FROM centos:7
COPY --from=build /go/src/github.com/openshift/eventrouter/eventrouter /bin/eventrouter
CMD ["/bin/eventrouter", "-v", "3", "-logtostderr"]
