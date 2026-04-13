#!/usr/bin/env python3
"""
kscript-api-map.py
Varre o repositorio Kora e produz um mapa completo da API KScript:
  - Modulos built-in registrados no checker
  - Funcoes emitidas no emitter
  - Objetos/components definidos em arquivos .ks de exemplos
  - Lifecycle hooks encontrados

Uso: python3 .claude/tools/kscript-api-map.py
Retorna JSON. Seguro: apenas leitura.
"""

import os
import re
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent.parent


def scan_go_for_builtins(base: Path) -> dict:
    """Procura mapa de builtins no compiler/checker."""
    builtins = {}
    checker_dir = base / "compiler" / "checker"
    if not checker_dir.exists():
        return builtins

    for f in checker_dir.rglob("*.go"):
        src = f.read_text(errors="replace")
        # Captura strings que parecem nomes de modulos built-in
        # Padrao: "NomeModulo" seguido de : ou { em contexto de mapa
        matches = re.findall(r'"([A-Z][a-zA-Z]+)"\s*[:{\/]', src)
        for m in matches:
            if m not in builtins:
                builtins[m] = {"source": str(f.relative_to(base))}

        # Captura funcoes associadas: "nomeFuncao"
        func_matches = re.findall(r'"([a-z][a-zA-Z]+)"\s*:', src)
        for fm in func_matches:
            if len(fm) > 2:  # filtrar chaves muito curtas
                builtins.setdefault("__functions__", []).append(fm)

    return builtins


def scan_emitter(base: Path) -> list:
    """Extrai cases de modulos built-in no emitter."""
    cases = []
    emitter_dir = base / "compiler" / "emitter"
    if not emitter_dir.exists():
        return cases

    for f in emitter_dir.rglob("*.go"):
        src = f.read_text(errors="replace")
        # case "NomeModulo":
        found = re.findall(r'case\s+"([A-Za-z][a-zA-Z0-9_]+)"', src)
        cases.extend(found)

    return sorted(set(cases))


def scan_ks_files(base: Path) -> dict:
    """Varre arquivos .ks em examples/ e templates/ buscando objects, components e hooks."""
    objects = []
    components = []
    hooks_found = set()
    ks_files = []

    LIFECYCLE_HOOKS = [
        "create", "update", "draw", "onDestroy",
        "onCollision", "onInput", "start", "destroy"
    ]

    for search_dir in ["examples", "templates"]:
        d = base / search_dir
        if not d.exists():
            continue
        for f in d.rglob("*.ks"):
            ks_files.append(str(f.relative_to(base)))
            src = f.read_text(errors="replace")

            # object NomeObjeto
            for m in re.findall(r'^\s*object\s+([A-Z][a-zA-Z0-9_]*)', src, re.MULTILINE):
                objects.append({"name": m, "file": str(f.relative_to(base))})

            # component NomeComponent
            for m in re.findall(r'^\s*component\s+([A-Z][a-zA-Z0-9_]*)', src, re.MULTILINE):
                components.append({"name": m, "file": str(f.relative_to(base))})

            # lifecycle hooks
            for hook in LIFECYCLE_HOOKS:
                pattern = rf'\b(async\s+)?{hook}\s*\('
                if re.search(pattern, src):
                    hooks_found.add(hook)

    return {
        "ks_files": ks_files,
        "objects": objects,
        "components": components,
        "lifecycle_hooks_used": sorted(hooks_found)
    }


def scan_kscript_types(base: Path) -> list:
    """Extrai tipos KScript referenciados no checker ou docs."""
    types_found = set()
    KNOWN_TYPES = [
        "bool", "int", "float", "string", "void",
        "Vector2", "Vector3", "Color", "Rect",
        "Array", "Map", "Entity", "Scene"
    ]
    checker_dir = base / "compiler" / "checker"
    if checker_dir.exists():
        for f in checker_dir.rglob("*.go"):
            src = f.read_text(errors="replace")
            for t in KNOWN_TYPES:
                if t in src:
                    types_found.add(t)
    return sorted(types_found)


def scan_core_modules(base: Path) -> list:
    """Lista modulos do core/ com seus pacotes Go."""
    core = base / "core"
    modules = []
    if not core.exists():
        return modules
    for d in sorted(core.iterdir()):
        if d.is_dir():
            go_files = list(d.glob("*.go"))
            test_files = list(d.glob("*_test.go"))
            modules.append({
                "name": d.name,
                "path": f"core/{d.name}",
                "go_files": len(go_files),
                "test_files": len(test_files),
                "has_tests": len(test_files) > 0
            })
    return modules


def main():
    result = {
        "tool": "kscript-api-map",
        "root": str(ROOT),
        "summary": {},
        "builtin_modules_checker": scan_go_for_builtins(ROOT),
        "builtin_cases_emitter": scan_emitter(ROOT),
        "kscript_types": scan_kscript_types(ROOT),
        "core_modules": scan_core_modules(ROOT),
        "ks_sources": scan_ks_files(ROOT)
    }

    # summary
    result["summary"] = {
        "core_module_count": len(result["core_modules"]),
        "core_modules_with_tests": sum(1 for m in result["core_modules"] if m["has_tests"]),
        "ks_file_count": len(result["ks_sources"]["ks_files"]),
        "objects_defined": len(result["ks_sources"]["objects"]),
        "components_defined": len(result["ks_sources"]["components"]),
        "emitter_builtin_cases": len(result["builtin_cases_emitter"]),
        "kscript_types_referenced": len(result["kscript_types"])
    }

    print(json.dumps(result, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    main()
