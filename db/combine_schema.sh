#!/bin/bash

# Combine Schema Script
# Automatically generates a complete DQL schema from all individual .dql files
# Outputs to schema/schema.dql with proper DQL syntax

echo "🔧 Combining all .dql schema files into complete schema..."
echo "========================================================"

# Base directory for schemas
SCHEMA_DIR="$(dirname "$0")/schema"
OUTPUT_FILE="$SCHEMA_DIR/schema.dql"

# Create backup of existing schema.dql if it exists
if [ -f "$OUTPUT_FILE" ]; then
    cp "$OUTPUT_FILE" "$OUTPUT_FILE.backup"
    echo "📋 Created backup: schema.dql.backup"
fi

# Initialize the output file with header
cat > "$OUTPUT_FILE" << 'EOF'
# Combined DQL Schema for DO Study LMS
# Auto-generated from individual .dql files
# Generated on: $(date)

EOF

echo "🔍 Scanning for .dql files..."
echo ""

# Counters
file_count=0
predicate_count=0
type_count=0

# Temporary files for organizing content
TEMP_PREDICATES="/tmp/predicates.tmp"
TEMP_TYPES="/tmp/types.tmp"

# Clear temp files
> "$TEMP_PREDICATES"
> "$TEMP_TYPES"

# Function to extract predicates from type definitions
extract_predicates_from_types() {
    local content="$1"
    local schema_name="$2"
    
    # Extract field definitions from type blocks and convert to predicates
    echo "$content" | awk -v schema="$schema_name" '
    /^type / { 
        in_type = 1
        type_name = $2
        next
    }
    /^}/ { 
        in_type = 0
        next
    }
    in_type && /^[[:space:]]*[a-zA-Z]/ {
        # Extract field name and type
        gsub(/^[[:space:]]+/, "")
        gsub(/[[:space:]]+$/, "")
        
        field_line = $0
        
        # Split on colon to get field name and type
        split(field_line, parts, ":")
        if (length(parts) >= 2) {
            field_name = parts[1]
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", field_name)
            
            type_part = parts[2]
            gsub(/^[[:space:]]+/, "", type_part)
            
            # Extract type and any directives
            if (match(type_part, /@/)) {
                type_only = substr(type_part, 1, RSTART-1)
                directives = substr(type_part, RSTART)
                gsub(/[[:space:]]+$/, "", type_only)
                gsub(/[[:space:]]+$/, "", directives)
                print field_name ": " type_only " " directives " ."
            } else {
                gsub(/[[:space:]]+$/, "", type_part)
                print field_name ": " type_part " ."
            }
        }
    }'
}

# Function to clean types (remove directives from field definitions)
clean_types() {
    local content="$1"
    
    echo "$content" | sed 's/@[^[:space:]]*//' | sed 's/[[:space:]]*$//' 
}

# Function to process a .dql file
process_dql_file() {
    local dql_file="$1"
    local schema_name="$2"
    
    if [ ! -s "$dql_file" ]; then
        echo "   ⚠️  Skipping $schema_name (empty file)"
        return
    fi
    
    echo "   📄 Processing $schema_name..."
    
    # Read the entire file content
    local file_content=$(cat "$dql_file")
    
    # Add comment headers
    echo "" >> "$TEMP_PREDICATES"
    echo "# $schema_name predicates" >> "$TEMP_PREDICATES"
    
    echo "" >> "$TEMP_TYPES"
    echo "# $schema_name types" >> "$TEMP_TYPES"
    
    # Extract predicates from type definitions
    extract_predicates_from_types "$file_content" "$schema_name" >> "$TEMP_PREDICATES"
    
    # Extract and clean type definitions
    echo "$file_content" | awk '
    /^type / { 
        in_type = 1
        print $0
        next
    }
    /^}/ { 
        if (in_type) {
            print $0
            in_type = 0
        }
        next
    }
    in_type {
        # Remove @index and other directives from type field definitions
        gsub(/@[^[:space:]]*/, "")
        # Remove periods and inline comments from type fields
        gsub(/[[:space:]]*\.[[:space:]]*#.*$/, "")
        gsub(/[[:space:]]*\.[[:space:]]*$/, "")
        gsub(/[[:space:]]+$/, "")
        if (length($0) > 0) print $0
    }' >> "$TEMP_TYPES"
    
    ((file_count++))
}

# Find and process all .dql files
find "$SCHEMA_DIR" -name "*.dql" -type f | grep -v "schema.dql" | grep -v "combined.dql" | sort | while read -r dql_file; do
    # Get relative path for naming
    rel_path=$(echo "$dql_file" | sed "s|$SCHEMA_DIR/||")
    schema_name=$(echo "$rel_path" | sed 's|/|_|g' | sed 's|\.dql$||')
    
    process_dql_file "$dql_file" "$schema_name"
done

# Wait for the subshell to complete and get the counts
file_count=$(find "$SCHEMA_DIR" -name "*.dql" -type f | grep -v "schema.dql" | grep -v "combined.dql" | wc -l | tr -d ' ')

# Add predicates section to output file
echo "" >> "$OUTPUT_FILE"
echo "# ============================================" >> "$OUTPUT_FILE"
echo "# PREDICATES (with indexes and types)" >> "$OUTPUT_FILE"
echo "# ============================================" >> "$OUTPUT_FILE"

if [ -s "$TEMP_PREDICATES" ]; then
    # Create a temporary file for smart deduplication
    TEMP_DEDUP="/tmp/predicates_dedup.tmp"
    
    # Remove comment lines and empty lines, then process predicates
    grep -v '^#' "$TEMP_PREDICATES" | grep -v '^[[:space:]]*$' | \
    awk '{
        # Extract predicate name (everything before the first colon)
        split($0, parts, ":")
        predicate_name = parts[1]
        gsub(/^[[:space:]]+|[[:space:]]+$/, "", predicate_name)
        
        # Store the full line for this predicate
        # If we already have this predicate, keep the one with more content (likely has @index)
        if (predicate_name in predicates) {
            # Keep the longer/more complex definition (usually the one with @index)
            if (length($0) > length(predicates[predicate_name])) {
                predicates[predicate_name] = $0
            }
        } else {
            predicates[predicate_name] = $0
        }
    }
    END {
        # Output all unique predicates, sorted by name
        for (name in predicates) {
            print predicates[name]
        }
    }' | sort >> "$OUTPUT_FILE"
fi

# Add types section to output file
echo "" >> "$OUTPUT_FILE"
echo "# ============================================" >> "$OUTPUT_FILE"
echo "# TYPES (structure definitions)" >> "$OUTPUT_FILE"
echo "# ============================================" >> "$OUTPUT_FILE"

if [ -s "$TEMP_TYPES" ]; then
    cat "$TEMP_TYPES" >> "$OUTPUT_FILE"
fi

# Clean up temp files
rm -f "$TEMP_PREDICATES" "$TEMP_TYPES"

# Update the date in the header
sed -i '' "s/Generated on: \$(date)/Generated on: $(date)/" "$OUTPUT_FILE"

echo ""
echo "✅ Schema combination completed!"
echo "================================================"
echo "📊 Summary:"
echo "   📁 Files processed: $file_count"
echo "   📄 Output file: $OUTPUT_FILE"
echo "   📏 Total lines: $(wc -l < "$OUTPUT_FILE" | tr -d ' ')"
echo ""
echo "🎯 Next steps:"
echo "   1. Review the generated schema.dql file"
echo "   2. Update deploy_all_schemas.sh to use schema.dql"
echo "   3. Deploy with: API_KEY=nZgKQjXX2XBRpt ./deploy_all_schemas.sh"
echo ""
echo "💡 Tip: Run this script whenever you modify individual .dql files"
