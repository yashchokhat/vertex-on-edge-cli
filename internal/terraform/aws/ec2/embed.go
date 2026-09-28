package ec2

import "embed"

//go:embed *.tf
var Templates embed.FS
