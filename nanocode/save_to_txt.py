import os
from pathlib import Path

BASE_DIR = Path(r"C:\Users\Administrator\Desktop\projects\claude-code\nanocode")
OUTPUT_FILE = BASE_DIR / "output.txt"

EXCLUDE_FILES = {"save_to_txt.py", "output.txt", "bun.lock", "package-lock.json"}
EXCLUDE_DIRS = {"dist", "dist-electron", "node_modules"}


def get_all_files():
    files = []
    for root, dirs, filenames in os.walk(BASE_DIR):
        root_path = Path(root)
        rel_root = root_path.relative_to(BASE_DIR)
        if rel_root != Path(".") and str(rel_root).split(os.sep)[0] in EXCLUDE_DIRS:
            continue
        for filename in filenames:
            if filename in EXCLUDE_FILES:
                continue
            if rel_root == Path("."):
                files.append(filename)
            else:
                files.append(str(rel_root / filename))
    return sorted(files, key=lambda x: (x.count(os.sep), x))


def main():
    all_files = get_all_files()

    with open(OUTPUT_FILE, "w", encoding="utf-8") as out:
        for rel_path in all_files:
            full_path = BASE_DIR / rel_path
            if full_path.exists():
                content = full_path.read_text(encoding="utf-8")
                out.write(rel_path + "\n")
                out.write(content + "\n")
                print(f"Added: {rel_path}")

    print(f"\nDone! Output saved to: {OUTPUT_FILE}")

if __name__ == "__main__":
    main()