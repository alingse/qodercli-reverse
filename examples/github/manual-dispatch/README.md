name: Qoder Manual Dispatch

on:
  workflow_dispatch:
    inputs:
      prompt:
        description: 'Prompt for the agent'
        required: true

jobs:
  run:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - run: cd decompiled && go build -o /usr/local/bin/qodercli-reverse .
      - env:
          INPUT_PROMPT: ${{ inputs.prompt }}
          GITHUB_OUTPUT: ${{ github.workspace }}/out.txt
        run: qodercli-reverse run