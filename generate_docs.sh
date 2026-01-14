$(go env GOPATH)/bin/gomarkdoc ./teams --output docs/teams.md
$(go env GOPATH)/bin/gomarkdoc ./channels --output docs/channels.md
$(go env GOPATH)/bin/gomarkdoc ./chats --output docs/chats.md
$(go env GOPATH)/bin/gomarkdoc ./models --output docs/models.md
$(go env GOPATH)/bin/gomarkdoc ./config --output docs/config.md
$(go env GOPATH)/bin/gomarkdoc ./search --output docs/search.md
$(go env GOPATH)/bin/gomarkdoc . --output docs/client.md