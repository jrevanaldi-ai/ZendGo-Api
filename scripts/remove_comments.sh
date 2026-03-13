#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

total_files=0
total_lines=0
removed_lines=0

remove_comments() {
    local file="$1"
    local temp_file=$(mktemp)
    
    while IFS= read -r line || [[ -n "$line" ]]; do
        if [[ -z "$line" ]]; then
            echo "" >> "$temp_file"
            continue
        fi
        
        if [[ "$line" =~ ^[[:space:]]*//[[:space:]]*$ ]]; then
            echo "$line" >> "$temp_file"
            continue
        fi
        
        if [[ "$line" =~ ^[[:space:]]*//go: ]] || \
           [[ "$line" =~ ^[[:space:]]*//sys ]] || \
           [[ "$line" =~ ^[[:space:]]*//go:build ]] || \
           [[ "$line" =~ ^[[:space:]]*//extern ]]; then
            echo "$line" >> "$temp_file"
            continue
        fi
        
        if [[ "$line" =~ https?:// ]]; then
            echo "$line" >> "$temp_file"
            continue
        fi
        
        if [[ "$line" =~ ^[[:space:]]*//.*package[[:space:]] ]]; then
            echo "$line" >> "$temp_file"
            continue
        fi
        
        if [[ "$line" =~ ^[[:space:]]*//.* ]] && \
           ! [[ "$line" =~ \" ]] && \
           ! [[ "$line" =~ \` ]]; then
            ((removed_lines++)) || true
            continue
        fi
        
        if [[ "$line" =~ // ]] && ! [[ "$line" =~ ^[[:space:]]*// ]]; then
            local in_string=false
            local in_rune=false
            local result=""
            local i=0
            local len=${#line}
            
            while [ $i -lt $len ]; do
                local char="${line:$i:1}"
                local next_char="${line:$((i+1)):1}"
                
                if [[ "$char" == "\\" ]]; then
                    result+="$char"
                    ((i++)) || true
                    if [ $i -lt $len ]; then
                        result+="${line:$i:1}"
                    fi
                    ((i++)) || true
                    continue
                fi
                
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
                
                if [[ "$char" == '`' ]]; then
                    result+="$char"
                    ((i++)) || true
                    continue
                fi
                
                if [[ "$char" == "/" ]] && [[ "$next_char" == "/" ]] && \
                   [[ "$in_string" == false ]] && [[ "$in_rune" == false ]]; then
                    break
                fi
                
                result+="$char"
                ((i++)) || true
            done
            
            result=$(echo "$result" | sed 's/[[:space:]]*$//')
            
            if [[ -n "$result" ]] && ! [[ "$result" =~ ^[[:space:]]*$ ]]; then
                echo "$result" >> "$temp_file"
            else
                ((removed_lines++)) || true
            fi
        else
            echo "$line" >> "$temp_file"
        fi
        
        ((total_lines++)) || true
        
    done < "$file"
    
    mv "$temp_file" "$file"
}

echo -e "${GREEN}=== Go Comment Remover ===${NC}"
echo ""

if [ ! -d "$1" ]; then
    echo -e "${RED}Error: Directory not found: $1${NC}"
    echo "Usage: $0 <directory>"
    exit 1
fi

TARGET_DIR="$1"

echo -e "${YELLOW}Scanning Go files in: $TARGET_DIR${NC}"
echo ""

while IFS= read -r -d '' file; do
    if [[ "$file" == *"/vendor/"* ]] || \
       [[ "$file" == *"/.git/"* ]] || \
       [[ "$file" == *"/testdata/"* ]]; then
        continue
    fi
    
    before_lines=$(wc -l < "$file")
    
    if [ "$before_lines" -eq 0 ]; then
        continue
    fi
    
    ((total_files++)) || true
    
    remove_comments "$file"
    
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
