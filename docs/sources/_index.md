---
title: Grafana Docs MCP Server
menuTitle: Grafana Docs MCP Server
description: Grafana Docs MCP Server is an MCP server that lets AI agents search, outline, and fetch Grafana Labs documentation from grafana.com/docs.
weight: 1
topicType: introduction
versionDate: 2026-09-01
cards:
  items:
    - title: Overview
      description: What the server does, who it's for, and how retrieval works.
      href: overview/
      height: 24
    - title: Cost
      description: The server adds no inference cost. You pay the agent's tokens.
      href: cost/
      height: 24
    - title: Get started
      description: Search, outline, and fetch docs from the CLI.
      href: get-started/
      height: 24
    - title: Install and connect
      description: Build the server and add it to Cursor, Claude Desktop, or Claude Code.
      href: install/
      height: 24
    - title: Tools reference
      description: Parameters, CLI flags, and a search-outline-fetch example.
      href: tools/
      height: 24
    - title: Configure the server
      description: DOCS_INDEX_URL, rate limits, and the URL allowlist.
      href: configure/
      height: 24
    - title: Integrate the core library
      description: Import pkg/grafanadocs in your own Go project.
      href: integrate/
      height: 24
  title_class: pt-0 lh-1
hero:
  title: Grafana Docs MCP Server
  description: Grafana Docs MCP Server is an MCP server that lets AI agents search, outline, and fetch Grafana Labs documentation from grafana.com/docs.
  level: 1
  height: 110
---

{{< docs/hero-simple key="hero" >}}

## Overview

AI agents answer from training data. That cutoff can be months old, incomplete, or wrong for the version you're running.

Grafana Docs MCP Server is a [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server. During a conversation, the agent searches `grafana.com/docs/`, reads the heading outline of a page, and fetches the section it needs as cleaned Markdown, with a source URL it can cite.

It retrieves sections rather than whole pages, so token use stays low. The server doesn't run an LLM, so it adds no inference cost. It serves current published documentation only.

The same retrieval ships in [gcx](https://github.com/grafana/gcx) (`gcx docs`) and [mcp-grafana](https://github.com/grafana/mcp-grafana). Use this project when you want a standalone MCP server that exposes only the docs tools.

## Explore

{{< card-grid key="cards" type="simple" >}}
