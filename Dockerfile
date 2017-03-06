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

# TODO - have a build distro machine. 
FROM fedora:latest
MAINTAINER Timothy St. Clair "tstclair@heptio.com"  

#TODO: RUN dnf update 
COPY eventrouter /eventrouter 

#TODO mount in config file and set defaults to parse /etc/$NAME
ENTRYPOINT ["/eventrouter"]

