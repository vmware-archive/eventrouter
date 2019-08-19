# Copyright 2017 Heptio Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
FROM golang:1.12.7 AS build

MAINTAINER Timothy St. Clair "tstclair@heptio.com"  
WORKDIR /src/
COPY . .

# ENTRYPOINT [ "/bin/sh", "-c" ]
# FROM base AS build

RUN CGO_ENABLED=0 go build -mod=vendor -o /src/eventrouter

FROM alpine:3.9 as release

WORKDIR /app
RUN apk update --no-cache && apk add ca-certificates
COPY --from=build /src/eventrouter .
USER nobody:nobody

ENTRYPOINT [ "./eventrouter" ]
CMD ["-v", "3", "-logtostderr"]