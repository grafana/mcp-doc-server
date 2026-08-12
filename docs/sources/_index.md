---
title: Grafana Docs MCP Server
menuTitle: Grafana Docs MCP Server
description: Grafana Docs MCP Server is an MCP server that gives AI agents live access to Grafana Labs product documentation instead of relying on stale training data.
weight: 1
topicType: introduction
versionDate: 2026-06-25
cards:
  items:
    - title: Overview
      description: Learn what Grafana Docs MCP Server does, when to use it, and how it works.
      href: overview/
      height: 24
    - title: Cost
      description: How costs break down. The server adds no inference cost, so you pay only for the agent's tokens.
      href: cost/
      height: 24
    - title: Get started
      description: Run the CLI to search, outline, and fetch Grafana documentation.
      href: get-started/
      height: 24
    - title: Install and connect
      description: Build the server and connect it to Cursor, Claude Desktop, or Claude Code.
      href: install/
      height: 24
    - title: Tools reference
      description: Tool parameters, CLI usage, and a full workflow example.
      href: tools/
      height: 24
    - title: Configure the server
      description: Environment variables, built-in limits, and the URL allowlist.
      href: configure/
      height: 24
    - title: Integrate the core library
      description: Embed the pkg/grafanadocs core in your own Go project.
      href: integrate/
      height: 24
  title_class: pt-0 lh-1
hero:
  title: Grafana Docs MCP Server
  description: Give AI agents live access to Grafana Labs product documentation instead of relying on stale training data.
  level: 1
  height: 110
---

{{< docs/hero-simple key="hero" >}}

## Overview

Give your AI agent live, citable Grafana Labs documentation during a conversation instead of stale training data. Grafana Docs MCP Server is a Model Context Protocol (MCP) server: agents search, browse, and retrieve current documentation from `grafana.com/docs/` as they answer.

Because the server retrieves individual sections rather than entire pages, agents get exactly the content they need at a low token cost, with citable source URLs. The server runs no large language model (LLM) of its own, so it adds no inference cost.

For what the server does, when to use it, and how retrieval works, refer to the [Overview](overview/).

## Get started

{{< card-grid key="cards" type="simple" >}}
