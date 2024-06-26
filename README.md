# Akeyless UI Webhook Builder

Akeyless UI Webhook Builder is a CLI tool that processes user flow recordings captured by the Chrome DevTools recorder and creates scripts that can be used in Akeyless custom rotator webhooks.

## Installation

To install Akeyless UI Webhook Builder, you need to have Go installed on your system. Then, you can use the following command:

```bash
go install github.com/akeyless-community/akeyless-ui-webhook-builder@latest
```

## Usage

To use Akeyless UI Webhook Builder, run the following command:

```bash
akeyless-ui-webhook-builder -f path/to/your/recording.json
```

The tool will guide you through the process of mapping the recorded steps to the necessary fields for password rotation.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

Apache 2.0
