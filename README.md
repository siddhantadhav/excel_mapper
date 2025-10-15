# 📊 excel_mapper

[![Go Reference](https://pkg.go.dev/badge/github.com/siddhantadhav/excel_mapper.svg)](https://pkg.go.dev/github.com/siddhantadhav/excel_mapper)  
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A **reusable Go package** to map, transform, and transfer data between Excel files — with support for direct column mappings, derived (computed) mappings, and raw formula-based transformations.  
Mappings can also be **stored and retrieved from MongoDB** for reuse or sharing.

---

## 🚀 Features

- Initialize and parse Excel files (`.xlsx`)
- Direct column-to-column mappings
- Derived mappings using custom transform functions
- Raw calculations (formula-based)
- MongoDB integration for persistent mapping storage
- Extensible design for future storage backends

---

## 🧩 Installation

```bash
go get github.com/siddhantadhav/excel_mapper
