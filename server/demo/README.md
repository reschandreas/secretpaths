# what is this?

Playing around with visualizing policies and secrets is no fun if you do not have any 
policies or secrets to visualize. This directory contains a small demo application
that generates random secrets and policies that can be used to test the visualization.

# How to use

Start a vault server in demo mode:

```shell
vault server -dev -dev-root-token-id="my-token"
```

Then run the demo application:

```shell
go run main.go
```