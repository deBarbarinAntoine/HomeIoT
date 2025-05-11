# Production Deployment

## Set the environment variables for the systemd service

Times change and so do best practices.

The current best way to do this is to `run systemctl edit myservice`, which will create an override file for you or let you edit an existing one.

In normal installations this will create a directory `/etc/systemd/system/myservice.service.d`, and inside that directory create a file whose name ends in `.conf` (typically, `override.conf`), and in this file you can add to or override any part of the unit shipped by the distribution.

For instance, in a file `/etc/systemd/system/myservice.service.d/myenv.conf`:

```toml
[Service]
Environment="SECRET=pGNqduRFkB4K9C2vijOmUDa2kPtUhArN"
Environment="ANOTHER_SECRET=JP8YLOc2bsNlrGuD6LVTq7L36obpjzxd"
```

Also note that if the directory exists and is empty, your service will be disabled! If you don't intend to put something in the directory, ensure that it does not exist.