#!/bin/bash

# Script untuk menghapus komentar // dari file Go
# Script ini hati-hati dan tidak akan menghapus:
# - Komentar dalam string ("..." atau `...`)
# - URL yang mengandung //
# - Package comment yang penting
# - Direktif compiler (//go:, //sys, dll)

set -e

# Warna untuk output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counter
total_files=0
total_lines=0
removed_lines=0

# Fungsi untuk menghapus komentar dengan aman
remove_comments() {
    local file="$1"
    local temp_file=$(mktemp)
    
    # Baca file line by line
    while IFS= read -r line || [[ -n "$line" ]]; do
        # Skip jika line kosong
        if [[ -z "$line" ]]; then
            echo "" >> "$temp_file"
            continue
        fi
        
        # Jangan hapus jika line hanya berisi //
        if [[ "$line" =~ ^[[:space:]]*//[[:space:]]*$ ]]; then
            echo "$line" >> "$temp_file"
            continue
        fi
        
        # Jangan hapus direktif compiler (//go:, //sys, //go:build, dll)
        if [[ "$line" =~ ^[[:space:]]*//go: ]] || \
           [[ "$line" =~ ^[[:space:]]*//sys ]] || \
           [[ "$line" =~ ^[[:space:]]*//go:build ]] || \
           [[ "$line" =~ ^[[:space:]]*//extern ]]; then
            echo "$line" >> "$temp_file"
            continue
        fi
        
        # Jangan hapus jika ada URL (http:// atau https://)
        if [[ "$line" =~ https?:// ]]; then
            echo "$line" >> "$temp_file"
            continue
        fi
        
        # Jangan hapus jika komentar ada di awal package declaration
        if [[ "$line" =~ ^[[:space:]]*//.*package[[:space:]] ]]; then
            echo "$line" >> "$temp_file"
            continue
        fi
        
        # Cek apakah line adalah pure comment (hanya komentar, tidak ada code)
        if [[ "$line" =~ ^[[:space:]]*//.* ]] && \
           ! [[ "$line" =~ \" ]] && \
           ! [[ "$line" =~ \` ]]; then
            # Ini adalah pure comment, skip (hapus)
            ((removed_lines++)) || true
            continue
        fi
        
        # Untuk line yang ada code + comment inline
        # Kita perlu hati-hati dengan string
        if [[ "$line" =~ // ]] && ! [[ "$line" =~ ^[[:space:]]*// ]]; then
            # Ada kemungkinan comment inline
            # Cek apakah // ada di dalam string
            local in_string=false
            local in_rune=false
            local result=""
            local i=0
            local len=${#line}
            
            while [ $i -lt $len ]; do
                local char="${line:$i:1}"
                local next_char="${line:$((i+1)):1}"
                
                # Handle escape character
                if [[ "$char" == "\\" ]]; then
                    result+="$char"
                    ((i++)) || true
                    if [ $i -lt $len ]; then
                        result+="${line:$i:1}"
                    fi
                    ((i++)) || true
                    continue
                fi
                
                # Handle string literal "
                if [[ "$char" == '"' ]] && [[ "$in_rune" == false ]]; then
                    if [[ "$in_string" == true ]]; then
                        in_string=false
                    else
                        in_string=true
                    fi
                    result+="$char"
                    ((i++)) || true
                    continue
                fi
                
                # Handle rune literal '
                if [[ "$char" == "'" ]] && [[ "$in_string" == false ]]; then
                    if [[ "$in_rune" == true ]]; then
                        in_rune=false
                    else
                        in_rune=true
                    fi
                    result+="$char"
                    ((i++)) || true
                    continue
                fi
                
                # Handle raw string `
                if [[ "$char" == '`' ]]; then
                    result+="$char"
                    ((i++)) || true
                    continue
                fi
                
                # Cek komentar //
                if [[ "$char" == "/" ]] && [[ "$next_char" == "/" ]] && \
                   [[ "$in_string" == false ]] && [[ "$in_rune" == false ]]; then
                    # Found comment, stop here
                    break
                fi
                
                result+="$char"
                ((i++)) || true
            done
            
            # Trim trailing whitespace
            result=$(echo "$result" | sed 's/[[:space:]]*$//')
            
            # Hanya tambahkan jika ada code
            if [[ -n "$result" ]] && ! [[ "$result" =~ ^[[:space:]]*$ ]]; then
                echo "$result" >> "$temp_file"
            else
                ((removed_lines++)) || true
            fi
        else
            # Tidak ada komentar, tambahkan line apa adanya
            echo "$line" >> "$temp_file"
        fi
        
        ((total_lines++)) || true
        
    done < "$file"
    
    # Ganti file original dengan hasil
    mv "$temp_file" "$file"
}

# Main script
echo -e "${GREEN}=== Go Comment Remover ===${NC}"
echo ""

# Cek apakah ada file Go
if [ ! -d "$1" ]; then
    echo -e "${RED}Error: Directory not found: $1${NC}"
    echo "Usage: $0 <directory>"
    exit 1
fi

TARGET_DIR="$1"

echo -e "${YELLOW}Scanning Go files in: $TARGET_DIR${NC}"
echo ""

# Temukan semua file .go
while IFS= read -r -d '' file; do
    # Skip file dalam vendor atau .git
    if [[ "$file" == *"/vendor/"* ]] || \
       [[ "$file" == *"/.git/"* ]] || \
       [[ "$file" == *"/testdata/"* ]]; then
        continue
    fi
    
    # Hitung jumlah baris sebelum
    before_lines=$(wc -l < "$file")
    
    if [ "$before_lines" -eq 0 ]; then
        continue
    fi
    
    ((total_files++)) || true
    
    # Proses file
    remove_comments "$file"
    
    # Hitung jumlah baris sesudah
    after_lines=$(wc -l < "$file")
    diff=$((before_lines - after_lines))
    
    if [ $diff -gt 0 ]; then
        echo -e "${GREEN}✓${NC} $(basename $file): Removed $diff comment lines"
    fi
    
done < <(find "$TARGET_DIR" -name "*.go" -type f -print0)

echo ""
echo -e "${GREEN}=== Summary ===${NC}"
echo -e "Total files processed: ${GREEN}$total_files${NC}"
echo -e "Total lines removed: ${YELLOW}$removed_lines${NC}"
echo ""
echo -e "${GREEN}Done!${NC}"
