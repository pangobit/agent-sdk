# MCP, but not dumb

## Introduction
This library is an opinionated take on MCP. More specifically, it provides MCP compatibility, but untangles the unnecesary jargon into a simple set of APIs that rely on standard API design pratcies. 

"Model Context Protocol" is probably the least descriptive possible name for what the standard is supposed to do. It is not strictly related to models or model context, and is not a protocol (but does specify protocols to use). 

Ultimately, MCP, in practice, is a domain-specific API schema for LLM-dependent applications. It prescribes an RPC message format (JSON-RPC 2.0), required endpoints, and a few data structures representing "tools" and other resources.
