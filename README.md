# traverser

Work with config data unmarshalled in to maps without knowing the shape of the data.

Used by:
- [envelope](https://github.com/mikesimons/envelope) to do recursive decryption and arbitrary key getting / setting.
- [kpatch](https://github.com/mikesimons/kpatch) to apply expression results to Kubernetes manifests.

Tested with data from common yaml, json and toml parsers.

Supports yaml.v3 yaml.Node type for yaml manipulation preserving some formatting / comments.
