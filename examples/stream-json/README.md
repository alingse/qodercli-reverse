# Stream JSON Example

`qodercli-reverse --output-format stream-json -p "hello"` emits newline-delimited JSON events. Each line is a JSON object with a `type` field (`start`, `text`, `tool_call`, `tool_result`, `finish`, `error`). Diagnostics go to stderr.