---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "sonatypeiq Provider"
subcategory: ""
description: |-
  
---

# sonatypeiq Provider



## Example Usage

```terraform
# 
# Copyright (c) 2019-present Sonatype, Inc.
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

provider "sonatypeiq" {
  host     = "my-sonatype-iq-server.tld:port"
  username = "username"
  password = "password"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `password` (String, Sensitive) Password for your Administrator user for Sonatype IQ Server
- `url` (String) Sonatype IQ Server URL
- `username` (String) Administrator Username for Sonatype IQ Server
