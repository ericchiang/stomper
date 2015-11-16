package main

import "testing"

func TestMerge(t *testing.T) {
	_, err := merge(newConfig, oldConfig)
	if err != nil {
		t.Fatal(err)
	}
}

var oldConfig = []byte(`{
  "id": "5a4526e952f0aa24f3fcc1b6971f7744eb5465d572a48d47c492cb6bbf9cbcda",
  "parent": "99fcaefe76ef1aa4077b90a413af57fd17d19dce4e50d7964a273aae67055235",
  "created": "2015-10-22T21:57:04.359313793Z",
  "container": "f915cef1707093c3f6884d11231561dfe7d36d9529f7a1be6f0613d0b6b64fd6",
  "container_config": {
    "Hostname": "f813a028846a",
    "Domainname": "",
    "User": "",
    "AttachStdin": false,
    "AttachStdout": false,
    "AttachStderr": false,
    "Tty": false,
    "OpenStdin": false,
    "StdinOnce": false,
    "Env": null,
    "Cmd": [
      "/bin/sh",
      "-c",
      "sed -i 's/^#\\s*\\(deb.*universe\\)$/\\1/g' /etc/apt/sources.list"
    ],
    "Image": "9e19ac89d27c13ef5acad3fd0d7c642e7d58ffd259913a9fd7665bf578444b5e",
    "Volumes": null,
    "WorkingDir": "",
    "Entrypoint": null,
    "OnBuild": null,
    "Labels": null
  },
  "docker_version": "1.8.2",
  "config": {
    "Hostname": "f813a028846a",
    "Domainname": "",
    "User": "",
    "AttachStdin": false,
    "AttachStdout": false,
    "AttachStderr": false,
    "Tty": false,
    "OpenStdin": false,
    "StdinOnce": false,
    "Env": null,
    "Cmd": null,
    "Image": "9e19ac89d27c13ef5acad3fd0d7c642e7d58ffd259913a9fd7665bf578444b5e",
    "Volumes": null,
    "WorkingDir": "",
    "Entrypoint": null,
    "OnBuild": null,
    "Labels": null
  },
  "architecture": "amd64",
  "os": "linux",
  "Size": 1895,
  "parent_id": "sha256:99fcaefe76ef1aa4077b90a413af57fd17d19dce4e50d7964a273aae67055235",
  "layer_id": "sha256:b9583a207297925b186c4e2f573f910b76e162804cf239df00ee2369d5779cf9"
}
`)

var newConfig = []byte(`{
  "id": "1d073211c498fd5022699b46a936b4e4bdacb04f637ad64d3475f558783f5c3e",
  "parent": "5a4526e952f0aa24f3fcc1b6971f7744eb5465d572a48d47c492cb6bbf9cbcda",
  "created": "2015-10-22T21:57:04.876735434Z",
  "container": "09d03ee5015ac389a9f4753345fe81bd846cc0813c795d28b1d5f60ac1fd38a7",
  "container_config": {
    "Hostname": "f813a028846a",
    "Domainname": "",
    "User": "",
    "AttachStdin": false,
    "AttachStdout": false,
    "AttachStderr": false,
    "Tty": false,
    "OpenStdin": false,
    "StdinOnce": false,
    "Env": null,
    "Cmd": [
      "/bin/sh",
      "-c",
      "#(nop) CMD [\"/bin/bash\"]"
    ],
    "Image": "ac65c371c3a545a83bfd46bfe1a2f304f85e3bc0f3ed0bc5922fcf6d3edd31be",
    "Volumes": null,
    "WorkingDir": "",
    "Entrypoint": null,
    "OnBuild": null,
    "Labels": null
  },
  "docker_version": "1.8.2",
  "config": {
    "Hostname": "f813a028846a",
    "Domainname": "",
    "User": "",
    "AttachStdin": false,
    "AttachStdout": false,
    "AttachStderr": false,
    "Tty": false,
    "OpenStdin": false,
    "StdinOnce": false,
    "Env": null,
    "Cmd": [
      "/bin/bash"
    ],
    "Image": "ac65c371c3a545a83bfd46bfe1a2f304f85e3bc0f3ed0bc5922fcf6d3edd31be",
    "Volumes": null,
    "WorkingDir": "",
    "Entrypoint": null,
    "OnBuild": null,
    "Labels": null
  },
  "architecture": "amd64",
  "os": "linux",
  "parent_id": "sha256:5a4526e952f0aa24f3fcc1b6971f7744eb5465d572a48d47c492cb6bbf9cbcda",
  "layer_id": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
}`)
