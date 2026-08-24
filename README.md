# 🎯 Offensive Tooling & Arsenal

Custom tooling, reconnaissance automation, and utility scripts engineered for offensive security operations. Focused on custom development, speed, and efficiency during engagements, not recycled scripts.

> ⚠️ **For educational and authorized-engagement use only.** Everything here is built for research, personal portfolio demonstration, and authorized testing. Unauthorized use against systems you do not own or have explicit permission to test is strictly prohibited.

---

## 📖 What's in here

| Tool / Directory | Language | Description |
|---|---|---|
| **`recon/`** | `Go / Python` | Automated reconnaissance utilities, target profiling, and infrastructure mapping scripts |
| **`network/`** | `Go` | Custom protocol wrappers, persistent session handlers, and concurrency-focused network tools |
| **`utils/`** | `Python / Bash` | Quick-parsing helpers, artifact extractors, and workflow accelerators |

---

## 🧠 Philosophy

Building tools isn't about collecting scripts — it's about **solving friction points in the engagement workflow**. This repo leans toward:

- **Zero Bloat:** Lightweight, purpose-built utilities designed with minimal external dependencies
- **Concurrency First:** Leveraging modern backend capabilities (like Go routines) to handle heavy enumeration workloads quickly
- **Operator-Driven:** Tools built to address specific bottlenecks encountered during reconnaissance and offensive operations

---

## 🗂️ Structure

```text
offensive-tooling/
├── recon/          # Custom reconnaissance & enumeration tools
├── network/        # Networking utilities, TCP/protocol handlers
├── utils/          # Parsing, data extraction, and helper scripts
└── README.md
