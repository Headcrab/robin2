#!/usr/bin/env python3
import gzip
import json
import sys


def to_int(value):
    if value in ("", None):
        return None
    return int(value)


def to_float(value):
    if value in ("", None):
        return None
    return float(value)


def to_str(value):
    if value in ("", None):
        return None
    return str(value)


def main() -> int:
    if len(sys.argv) != 2:
        return 1

    with gzip.open(sys.argv[1], "rt", encoding="utf-8") as fh:
        payload = json.load(fh)

    if isinstance(payload, dict):
        payload = [payload]
    if not isinstance(payload, list):
        raise TypeError("truckscales payload must be a JSON object or array")

    for row in payload:
        dt = row.get("DateTime")
        if not dt:
            continue

        normalized = {
            "DateTime": dt,
            "InvNum": to_str(row.get("InvNum")),
            "VagNum": int(row["VagNum"]),
            "Brutto": int(row["Brutto"]),
            "Tare": to_int(row.get("Tare")),
            "Netto": to_int(row.get("Netto")),
            "Difference": to_int(row.get("Difference")),
            "Carrying": to_int(row.get("Carrying")),
            "Velocity": to_float(row.get("Velocity")),
            "Cargotype": to_str(row.get("Cargotype")),
        }
        print(json.dumps(normalized, ensure_ascii=False))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
