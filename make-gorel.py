#!/usr/bin/env python3
#
# SPDX-FileCopyrightText: 2025 Buoyant Inc.
# SPDX-License-Identifier: Apache-2.0
#
# Copyright 2025 Buoyant Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may
# not use this file except in compliance with the License.  You may obtain
# a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import sys

import yaml

OSES = [ "linux", "windows" ]
ARCHITECTURES = [ "arm64", "amd64" ]

class ImageStyle:
    def __init__(self, imagesuffix, dockerfile, build_args=None):
        self.imagesuffix = imagesuffix
        self.dockerfile = dockerfile
        self.build_args = build_args

class BuildStyle:
    def __init__(self, name, image_styles, oses=OSES, architectures=ARCHITECTURES):
        self.name = name
        self.oses = oses
        self.architectures = architectures
        self.image_styles = image_styles

def BuildStyleGUI(name):
    return BuildStyle("generic",
                      [
                        ImageStyle(name, "gui",
                                   [ "--build-arg=WORKLOAD=%s-workload" % name ]),
                      ])

def BuildStyleGeneric(name):
    return BuildStyle("generic",
                      [
                        ImageStyle(name, "workload",
                                   [ "--build-arg=WORKLOAD=%s-workload" % name ]),
                      ])

def BuildStyleExternal(name):
    return BuildStyle("generic",
                      [
                        ImageStyle(f"external-{name}", "external-workload",
                                   [ "--build-arg=EXTERNAL_BASE={{ .Env.EXTERNAL_BASE }}",
                                     "--build-arg=WORKLOAD=%s-workload" % name ]),
                        ImageStyle(f"bel-external-{name}", "bel-external-workload",
                                   [ "--build-arg=EXTERNAL_BASE={{ .Env.BEL_EXTERNAL_BASE }}",
                                     "--build-arg=WORKLOAD=%s-workload" % name ]),
                      ],
                      oses = [ "linux" ])


def BuildStylePi(name):
    return BuildStyle("pi",
                      [
                        ImageStyle(f"pi-{name}", "external-workload",
                                   [ "--build-arg=EXTERNAL_BASE={{ .Env.EXTERNAL_BASE }}",
                                     "--build-arg=WORKLOAD=%s-workload" % name ]),
                        ImageStyle(f"bel-pi-{name}", "bel-external-workload",
                                   [ "--build-arg=EXTERNAL_BASE={{ .Env.BEL_EXTERNAL_BASE }}",
                                     "--build-arg=WORKLOAD=%s-workload" % name ]),
                      ],
                      architectures=[ "arm64" ],
                      oses = [ "linux" ])

class Build:
    def __init__(self,
                 name,
                 build_styles=None,
                 extra_files=None,
    ):
        self.name = name
        self.build_styles = build_styles
        self.extra_files = extra_files

BUILDS = [
    Build("gui",
          build_styles=[ BuildStyleGUI("gui") ],
          extra_files=[ "assets/html" ],
    ),
    Build("load",
          build_styles=[ BuildStyleGeneric("load") ]),
    Build("face",
          build_styles=[ BuildStyleGeneric("face"),
                         BuildStyleExternal("face") ]),
    Build("color",
          build_styles=[ BuildStyleGeneric("color"),
                         BuildStyleExternal("color"),
                         BuildStylePi("color") ]),
    Build("smiley",
          build_styles=[ BuildStyleGeneric("smiley"),
                         BuildStyleExternal("smiley"),
                         BuildStylePi("smiley") ]),
]

build_defs = {}
docker_defs = []
manifest_defs = []
archive_info = {
    "generic-linux": {},
    "generic-windows": {},
    "pi-linux": {},
}
archive_defs = []

for build in BUILDS:
    for build_style in build.build_styles:
        build_name = build.name
        style_name = build_style.name

        for os in list(build_style.oses):
            build_id = f"{style_name}-{build_name}-{os}"

            build_def = {
                "id": build_id,
                "main": f"./cmd/{style_name}/{build_name}",
                "binary": f"{build_name}-workload",
                "env": [ "CGO_ENABLED=0" ],
                "goos": [ os ],
                "goarch": list(build_style.architectures),
            }

            build_defs[build_id] = build_def

            # print(f"# {build_name} {style_name} {list(build_style.oses)}")
            # print(f"# {build_name} {style_name} {os}: {build_def}")
            archive_id = f"{style_name}-{os}"
            archive_info[archive_id][build_id] = True

        for image in build_style.image_styles:
            dockerfile = f"Dockerfiles/Dockerfile.{image.dockerfile}"
            image_name = "{{ .Env.REGISTRY }}/{{ .Env.IMAGE_NAME }}-%s" % image.imagesuffix

            currents = []
            latests = []

            for arch in build_style.architectures:
                current_image = "%s:{{ .Version }}-%s" % (image_name, arch)
                latest_image = "%s:latest-%s" % (image_name, arch)

                currents.append(current_image)
                latests.append(latest_image)

                build_flags = [
                    f"--platform=linux/{arch}",
                ]

                if image.build_args:
                    build_flags.extend(image.build_args)

                build_id = f"{style_name}-{build_name}-linux"

                docker_def = {
                    "use": "buildx",
                    "goos": "linux",
                    "goarch": arch,
                    "dockerfile": dockerfile,
                    "ids": [ build_id ],
                    "image_templates": [ current_image, latest_image ],
                    "build_flag_templates": build_flags,
                }

                if build.extra_files:
                    docker_def["extra_files"] = list(build.extra_files)

                docker_defs.append(docker_def)

            current_manifest_name = "%s:{{ .Version }}" % image_name
            latest_manifest_name = "%s:latest" % image_name

            current_manifest_def = {
                "name_template": current_manifest_name,
                "image_templates": currents,
                "create_flags": [ "--insecure" ],
                "push_flags": [ "--insecure" ],
            }

            latest_manifest_def = {
                "name_template": latest_manifest_name,
                "image_templates": latests,
                "create_flags": [ "--insecure" ],
                "push_flags": [ "--insecure" ],
            }

            manifest_defs.append(current_manifest_def)
            manifest_defs.append(latest_manifest_def)

# print("#")

for archive_name, build_ids in archive_info.items():
    archive_id = archive_name
    os = "linux"

    if "-" in archive_name:
        archive_id, os = archive_name.split("-")

    name_template = '{{ .ProjectName }}_%s_{{ .Version }}_%s_{{ .Arch }}{{ with .Arm }}v{{ . }}{{ end }}{{ with .Mips }}_{{ . }}{{ end }}{{ if not (eq .Amd64 "v1") }}{{ .Amd64 }}{{ end }}'
    name_template = name_template % (archive_id, os)

    build_ids = sorted(archive_info[archive_name].keys())
    # print(f"# {archive_id}-{os}: {build_ids}")

    archive_def = {
        "id": archive_name,
        "name_template": name_template,
        "builds": build_ids,
        "files": [ "none*" ],
        "allow_different_binary_count": True,
    }

    # Windows archives should be zip format
    if os == "windows":
        archive_def["format"] = "zip"

    archive_defs.append(archive_def)

gorel = {
    "builds": list(build_defs.values()),
    "dockers": docker_defs,
    "docker_manifests": manifest_defs,
    "archives": archive_defs,
}

print(sys.stdin.read())
print(yaml.safe_dump(gorel))

