import os

def collect_go_files():
    output_file = 'code.txt'
    
    with open(output_file, 'w', encoding='utf-8') as outfile:
        for root, dirs, files in os.walk('.'):
            # Пропускаем скрытые директории и vendor
            dirs[:] = [d for d in dirs if not d.startswith('.') and d != 'vendor']
            
            for file in files:
                if file.endswith('.go'):
                    file_path = os.path.join(root, file)
                    # Делаем путь относительным (убираем ./ в начале если есть)
                    rel_path = os.path.relpath(file_path, '.')
                    
                    try:
                        with open(file_path, 'r', encoding='utf-8') as infile:
                            content = infile.read()
                            
                        # Формат: относительный путь\nкод\n\n
                        outfile.write(f"{rel_path}\n{content}\n\n")
                    except Exception as e:
                        print(f"Ошибка при чтении {rel_path}: {e}")

    print(f"Готово! Все .go файлы собраны в {output_file}")

if __name__ == "__main__":
    collect_go_files()
