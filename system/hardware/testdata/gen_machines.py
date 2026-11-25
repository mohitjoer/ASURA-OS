#!/usr/bin/env python3
import json
import os
import shutil

ROOT = os.path.dirname(os.path.abspath(__file__))
MACHINES_DIR = os.path.join(ROOT, "machines")

machines = [
    {
        "id": "generic-laptop",
        "dmi": {
            "sys_vendor": "GenericVendor",
            "product_name": "Portable Laptop",
            "product_family": "Portable",
            "board_name": "GenericBoard",
            "chassis_type": "10",  # Notebook
            "bios_vendor": "GenericBIOS",
            "bios_version": "1.0.0",
            "bios_date": "01/01/2024",
            "product_serial": "[REDACTED]",
            "chassis_serial": "[REDACTED]",
            "board_serial": "[REDACTED]",
        },
        "pci": [
            {"address": "0000:00:00.0", "vendor": "0x8086", "device": "0x46a6", "class": "0x060000"}
        ],
        "usb": [
            {"name": "usb1", "idVendor": "1d6b", "idProduct": "0002"}
        ],
        "modules": ["snd_hda_intel"],
        "expected": ["generic-laptop", "intel-audio"]
    },
    {
        "id": "acer-laptop",
        "dmi": {
            "sys_vendor": "Acer Inc.",
            "product_name": "Aspire A515-56",
            "product_family": "Aspire",
            "board_name": "Jimny_TL",
            "chassis_type": "10",
            "bios_vendor": "Insyde",
            "bios_version": "V1.20",
            "bios_date": "02/15/2022",
            "product_serial": "[REDACTED]",
            "chassis_serial": "[REDACTED]",
            "board_serial": "[REDACTED]",
        },
        "pci": [
            {"address": "0000:00:00.0", "vendor": "0x8086", "device": "0x9a14", "class": "0x060000"}
        ],
        "usb": [],
        "modules": ["acer_wmi"],
        "expected": ["generic-laptop", "acer-laptop"]
    },
    {
        "id": "apple-macbook",
        "dmi": {
            "sys_vendor": "Apple Inc.",
            "product_name": "MacBookPro14,1",
            "product_family": "MacBook Pro",
            "board_name": "Mac-B4831CEBD52A0C4C",
            "chassis_type": "9",  # Laptop
            "bios_vendor": "Apple Inc.",
            "bios_version": "522.0.0.0.0",
            "bios_date": "10/27/2023",
            "product_serial": "[REDACTED]",
            "chassis_serial": "[REDACTED]",
            "board_serial": "[REDACTED]",
        },
        "pci": [],
        "usb": [],
        "modules": ["applesmc"],
        "expected": ["generic-laptop", "apple-macbook"]
    },
    {
        "id": "asus-laptop",
        "dmi": {
            "sys_vendor": "ASUSTeK COMPUTER INC.",
            "product_name": "ZenBook UX425EA",
            "product_family": "ZenBook",
            "board_name": "UX425EA",
            "chassis_type": "10",
            "bios_vendor": "American Megatrends Inc.",
            "bios_version": "UX425EA.311",
            "bios_date": "06/10/2021",
            "product_serial": "[REDACTED]",
            "chassis_serial": "[REDACTED]",
            "board_serial": "[REDACTED]",
        },
        "pci": [],
        "usb": [],
        "modules": ["asus_wmi"],
        "expected": ["generic-laptop", "asus-laptop"]
    },
    {
        "id": "asus-tuf-fx507zc4",
        "dmi": {
            "sys_vendor": "ASUSTeK COMPUTER INC.",
            "product_name": "ASUS TUF Gaming F15 FX507ZC4_FX507ZC4",
            "product_family": "ASUS TUF Gaming F15",
            "board_name": "FX507ZC4",
            "chassis_type": "10",
            "bios_vendor": "American Megatrends Inc.",
            "bios_version": "FX507ZC4.308",
            "bios_date": "11/24/2023",
            "product_serial": "[REDACTED]",
            "chassis_serial": "[REDACTED]",
            "board_serial": "[REDACTED]",
        },
        "pci": [
            {"address": "0000:01:00.0", "vendor": "0x10de", "device": "0x25a2", "class": "0x030000"}
        ],
        "usb": [],
        "modules": ["asus_wmi", "nvidia"],
        "expected": ["generic-laptop", "asus-laptop", "asus-tuf-fx507zc4", "nvidia-gpu"]
    },
    {
        "id": "dell-laptop",
        "dmi": {
            "sys_vendor": "Dell Inc.",
            "product_name": "Latitude 5420",
            "product_family": "Latitude",
            "board_name": "08YDFV",
            "chassis_type": "10",
            "bios_vendor": "Dell Inc.",
            "bios_version": "1.14.0",
            "bios_date": "08/12/2022",
            "product_serial": "[REDACTED]",
            "chassis_serial": "[REDACTED]",
            "board_serial": "[REDACTED]",
        },
        "pci": [],
        "usb": [],
        "modules": ["dell_wmi"],
        "expected": ["generic-laptop", "dell-laptop"]
    },
    {
        "id": "dell-xps",
        "dmi": {
            "sys_vendor": "Dell Inc.",
            "product_name": "XPS 13 9310",
            "product_family": "XPS",
            "board_name": "0DXP13",
            "chassis_type": "10",
            "bios_vendor": "Dell Inc.",
            "bios_version": "3.11.0",
            "bios_date": "05/18/2023",
            "product_serial": "[REDACTED]",
            "chassis_serial": "[REDACTED]",
            "board_serial": "[REDACTED]",
        },
        "pci": [],
        "usb": [],
        "modules": ["dell_wmi"],
        "expected": ["generic-laptop", "dell-laptop", "dell-xps"]
    },
    {
        "id": "framework-laptop",
        "dmi": {
            "sys_vendor": "Framework",
            "product_name": "Laptop 13 (AMD Ryzen 7040Series)",
            "product_family": "Laptop 13",
            "board_name": "FRANMDCP07",
            "chassis_type": "10",
            "bios_vendor": "INSYDE Corp.",
            "bios_version": "03.05",
            "bios_date": "03/29/2024",
            "product_serial": "[REDACTED]",
            "chassis_serial": "[REDACTED]",
            "board_serial": "[REDACTED]",
        },
        "pci": [],
        "usb": [],
        "modules": [],
        "expected": ["generic-laptop", "framework-laptop"]
    },
    {
        "id": "hp-laptop",
        "dmi": {
            "sys_vendor": "HP",
            "product_name": "HP EliteBook 840 G8",
            "product_family": "EliteBook",
            "board_name": "880D",
            "chassis_type": "10",
            "bios_vendor": "HP",
            "bios_version": "T76 Ver. 01.12.00",
            "bios_date": "01/10/2023",
            "product_serial": "[REDACTED]",
            "chassis_serial": "[REDACTED]",
            "board_serial": "[REDACTED]",
        },
        "pci": [],
        "usb": [],
        "modules": ["hp_wmi"],
        "expected": ["generic-laptop", "hp-laptop"]
    },
    {
        "id": "hp-omen-victus",
        "dmi": {
            "sys_vendor": "HP",
            "product_name": "OMEN by HP Laptop 16-b0xxx",
            "product_family": "Omen",
            "board_name": "88F4",
            "chassis_type": "10",
            "bios_vendor": "AMI",
            "bios_version": "F.15",
            "bios_date": "09/14/2022",
            "product_serial": "[REDACTED]",
            "chassis_serial": "[REDACTED]",
            "board_serial": "[REDACTED]",
        },
        "pci": [],
        "usb": [],
        "modules": ["hp_wmi"],
        "expected": ["generic-laptop", "hp-laptop", "hp-omen-victus"]
    },
    {
        "id": "lenovo-legion",
        "dmi": {
            "sys_vendor": "LENOVO",
            "product_name": "Legion 5 15ACH6H",
            "product_family": "Legion 5 15ACH6H",
            "board_name": "LNVNB161216",
            "chassis_type": "10",
            "bios_vendor": "LENOVO",
            "bios_version": "GKCN59WW",
            "bios_date": "10/12/2022",
            "product_serial": "[REDACTED]",
            "chassis_serial": "[REDACTED]",
            "board_serial": "[REDACTED]",
        },
        "pci": [],
        "usb": [],
        "modules": ["ideapad_laptop"],
        "expected": ["generic-laptop", "lenovo-legion"]
    },
    {
        "id": "lenovo-thinkpad",
        "dmi": {
            "sys_vendor": "LENOVO",
            "product_name": "ThinkPad T14 Gen 2",
            "product_family": "ThinkPad T14 Gen 2",
            "board_name": "20W0003KUS",
            "chassis_type": "10",
            "bios_vendor": "LENOVO",
            "bios_version": "N34ET56W (1.56 )",
            "bios_date": "08/30/2023",
            "product_serial": "[REDACTED]",
            "chassis_serial": "[REDACTED]",
            "board_serial": "[REDACTED]",
        },
        "pci": [],
        "usb": [
            {"name": "1-3", "idVendor": "06cb", "idProduct": "00bd"}
        ],
        "modules": ["thinkpad_acpi"],
        "expected": ["generic-laptop", "lenovo-thinkpad", "fprintd-synaptics"]
    }
]

def main():
    os.makedirs(MACHINES_DIR, exist_ok=True)
    for m in machines:
        dirpath = os.path.join(MACHINES_DIR, m["id"])
        if os.path.exists(dirpath):
            shutil.rmtree(dirpath)
        os.makedirs(dirpath, exist_ok=True)

        # DMI
        dmi_dir = os.path.join(dirpath, "class", "dmi", "id")
        os.makedirs(dmi_dir, exist_ok=True)
        for k, v in m["dmi"].items():
            with open(os.path.join(dmi_dir, k), "w") as f:
                f.write(v + "\n")

        # PCI
        pci_dir = os.path.join(dirpath, "bus", "pci", "devices")
        os.makedirs(pci_dir, exist_ok=True)
        lspci_lines = []
        for dev in m["pci"]:
            dev_dir = os.path.join(pci_dir, dev["address"])
            os.makedirs(dev_dir, exist_ok=True)
            with open(os.path.join(dev_dir, "vendor"), "w") as f:
                f.write(dev["vendor"] + "\n")
            with open(os.path.join(dev_dir, "device"), "w") as f:
                f.write(dev["device"] + "\n")
            with open(os.path.join(dev_dir, "class"), "w") as f:
                f.write(dev["class"] + "\n")
            lspci_lines.append(f"{dev['address']} Class {dev['class']}: {dev['vendor']}:{dev['device']}")

        with open(os.path.join(dirpath, "lspci.txt"), "w") as f:
            f.write("\n".join(lspci_lines) + "\n")

        # USB
        usb_dir = os.path.join(dirpath, "bus", "usb", "devices")
        os.makedirs(usb_dir, exist_ok=True)
        lsusb_lines = []
        for u in m["usb"]:
            u_dir = os.path.join(usb_dir, u["name"])
            os.makedirs(u_dir, exist_ok=True)
            with open(os.path.join(u_dir, "idVendor"), "w") as f:
                f.write(u["idVendor"] + "\n")
            with open(os.path.join(u_dir, "idProduct"), "w") as f:
                f.write(u["idProduct"] + "\n")
            lsusb_lines.append(f"Bus 001 Device 001: ID {u['idVendor']}:{u['idProduct']}")

        with open(os.path.join(dirpath, "lsusb.txt"), "w") as f:
            f.write("\n".join(lsusb_lines) + "\n")

        # Modules
        with open(os.path.join(dirpath, "modules.txt"), "w") as f:
            for mod in m["modules"]:
                f.write(f"{mod} 16384 0 - Live 0x0\n")

        # Expected layers
        with open(os.path.join(dirpath, "expected_layers.json"), "w") as f:
            json.dump(m["expected"], f, indent=2)

    print(f"Generated {len(machines)} machine fixtures in {MACHINES_DIR}")

if __name__ == "__main__":
    main()
