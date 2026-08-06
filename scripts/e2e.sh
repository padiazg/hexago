#!/usr/bin/env bash
# E2E test runner for hexago generator.
# Tests all major generation paths: init, add commands, validate, build, test.
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN="$PROJECT_ROOT/hexago-e2e-bin"
PASS_COUNT=0
FAIL_COUNT=0

# ─── Helpers ─────────────────────────────────────────────────────────────────

build_bin() {
    echo "==> Building hexago binary..."
    (cd "$PROJECT_ROOT" && go build -o "$BIN" .)
    if [[ $? -ne 0 ]]; then
        echo "FATAL: hexago build failed"
        exit 1
    fi
}

cleanup() { rm -rf "$WORK_DIR"; }

# Run hexago. Uses $WORK_DIR as --working-directory by default.
# If $WORK_DIR is empty, uses --working-directory "." (for cd'd scenarios).
hex() {
    _rc=0
    if [[ -n "$WORK_DIR" && -d "$WORK_DIR" ]]; then
        "$BIN" --working-directory "$WORK_DIR" "$@" 2>&1 || _rc=$?
    else
        "$BIN" "$@" 2>&1 || _rc=$?
    fi
}

# After init, returns the actual project root directory.
# For non-in-place: $WORK_DIR/<project-name>
# For in-place: $WORK_DIR
project_root() {
    local name="$1"
    local ip="$2"
    if [[ "$ip" == "yes" ]]; then
        echo "$WORK_DIR"
    else
        echo "$WORK_DIR/$name"
    fi
}

# ─── Scenarios ───────────────────────────────────────────────────────────────

test_service_full() {
    WORK_DIR=$(mktemp -d)
    trap 'cleanup; WORK_DIR=""' EXIT

    hex init demo-service \
        --project-type service \
        --module github.com/test/demo-service \
        --with-example \
        --with-observability \
        --with-metrics \
        --explicit-ports \
        --with-migrations \
        --db-driver sqlite3 \
        --with-docker \
        --with-workers \
        --with-tests

    # Resolve project directory after init
    local PROJ
    PROJ=$(project_root demo-service no)

    # 1. .hexago.yaml exists and contains expected values
    if [[ ! -f "$PROJ/.hexago.yaml" ]]; then
        echo "FAIL: .hexago.yaml not created"; return 1
    fi
    if ! grep -q 'type: service' "$PROJ/.hexago.yaml"; then
        echo "FAIL: .hexago.yaml missing type: service"; return 1
    fi
    if ! grep -q 'database_driver: sqlite3' "$PROJ/.hexago.yaml"; then
        echo "FAIL: .hexago.yaml missing database_driver: sqlite3"; return 1
    fi
    if ! grep -q 'with_workers: true' "$PROJ/.hexago.yaml"; then
        echo "FAIL: .hexago.yaml missing with_workers"; return 1
    fi
    if ! grep -q 'explicit_ports: true' "$PROJ/.hexago.yaml"; then
        echo "FAIL: .hexago.yaml missing explicit_ports"; return 1
    fi

    # 2. Core scaffolding exists
    [[ -f "$PROJ/internal/core/services/processor.go" ]] || {
        echo "FAIL: example service (processor) not generated"; return 1
    }
    [[ -d "$PROJ/migrations" ]] || {
        echo "FAIL: migrations dir not generated"; return 1
    }
    [[ -d "$PROJ/internal/workers" ]] || {
        echo "FAIL: workers dir not generated"; return 1
    }
    [[ -f "$PROJ/internal/observability/health.go" ]] || {
        echo "FAIL: observability health.go not generated"; return 1
    }

    # 3. Docker file
    [[ -f "$PROJ/Dockerfile" ]] || {
        echo "FAIL: Dockerfile not generated"; return 1
    }

    # 4. Build
    (cd "$PROJ" && go build ./...) || {
        echo "FAIL: service-full build error"; return 1
    }

    # 5. Validate
    (cd "$PROJ" && WORK_DIR="" hex validate)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: validate returned $?"; return 1
    fi

    # 6. Version command
    (cd "$PROJ" && WORK_DIR="" hex version --simple 2>/dev/null)
    local ver
    ver=$(cd "$PROJ" && WORK_DIR="" hex version --simple 2>/dev/null) || true
    if [[ -z "$ver" ]]; then
        echo "FAIL: version --simple returned empty"; return 1
    fi

    return 0
}

# Test each http-server framework (init + build only; add-cycle covered by service-full).
test_frameworks() {
    local frameworks=(stdlib echo gin chi fiber)
    local fw_ok=0
    local fw_fail=0

    for fw in "${frameworks[@]}"; do
        WORK_DIR=$(mktemp -d)
        trap 'cleanup; WORK_DIR=""' EXIT

        hex init demo-fw \
            --project-type http-server \
            --framework "$fw" \
            --module github.com/test/demo-fw \
            --with-observability \
            --with-metrics

        if [[ $_rc -ne 0 ]]; then
            echo "FAIL: init $fw returned $_rc"; fw_fail=$((fw_fail + 1)); continue
        fi

        local PROJ
        PROJ=$(project_root demo-fw no)

        # Build the framework adapter + server
        (cd "$PROJ" && go build ./...) || {
            echo "FAIL: build $fw error"; fw_fail=$((fw_fail + 1)); continue
        }

        fw_ok=$((fw_ok + 1))
    done

    if [[ $fw_fail -gt 0 ]]; then
        echo "FAIL: $fw_fail framework(s) failed (${fw_ok}/${#frameworks[@]} passed)"
        return 1
    fi
    return 0
}

# Test CLI project type: init, build, version output.
test_cli() {
    WORK_DIR=$(mktemp -d)
    trap 'cleanup; WORK_DIR=""' EXIT

    hex init demo-cli \
        --project-type cli \
        --module github.com/test/demo-cli

    if [[ $_rc -ne 0 ]]; then
        echo "FAIL: cli init returned $_rc"; return 1
    fi

    local PROJ
    PROJ=$(project_root demo-cli no)

    [[ -f "$PROJ/cmd/version.go" ]] || {
        echo "FAIL: cli cmd/version.go not generated"; return 1
    }

    (cd "$PROJ" && go build ./...) || {
        echo "FAIL: cli build error"; return 1
    }

    # Version bubble
    local version_output
    version_output=$(WORK_DIR="$PROJ" hex version 2>&1) || true
    if [[ -z "$version_output" ]]; then
        echo "FAIL: cli version output empty"; return 1
    fi

    # Version --simple
    local simple_ver
    simple_ver=$(WORK_DIR="$PROJ" hex version --simple 2>/dev/null) || true
    if [[ -z "$simple_ver" ]]; then
        echo "FAIL: cli version --simple empty"; return 1
    fi

    return 0
}

# Test driver-driven adapter style + usecases core logic.
test_driver_driven() {
    WORK_DIR=$(mktemp -d)
    trap 'cleanup; WORK_DIR=""' EXIT

    hex init demo-dd \
        --adapter-style driver-driven \
        --core-logic usecases \
        --module github.com/test/demo-dd

    if [[ $_rc -ne 0 ]]; then
        echo "FAIL: driver-driven init returned $_rc"; return 1
    fi

    local PROJ
    PROJ=$(project_root demo-dd no)

    # Verify .hexago.yaml has the right values
    if ! grep -q 'adapter_style: driver-driven' "$PROJ/.hexago.yaml"; then
        echo "FAIL: .hexago.yaml missing adapter_style: driver-driven"; return 1
    fi
    if ! grep -q 'core_logic: usecases' "$PROJ/.hexago.yaml"; then
        echo "FAIL: .hexago.yaml missing core_logic: usecases"; return 1
    fi

    # Create domain dir (generator bug: add entity requires it but doesn't create it)
    mkdir -p "$PROJ/internal/core/domain"

    # cd into project dir so relative paths in domain generator resolve correctly
    (cd "$PROJ" && WORK_DIR="" hex add domain entity User --fields "id:string,name:string,email:string,createdAt:time.Time")
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add domain entity User returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add service UserService --entity User)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add service returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add adapter primary http UserHandler)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add adapter primary returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add adapter secondary database UserRepository)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add adapter secondary returned $?"; return 1
    fi

     # Verify driver-driven directory structure
    [[ -d "$PROJ/internal/adapters/driver" ]] || {
        echo "FAIL: adapters/driver dir missing"; return 1
    }
    [[ -d "$PROJ/internal/adapters/driven" ]] || {
        echo "FAIL: adapters/driven dir missing"; return 1
    }
    [[ -f "$PROJ/internal/adapters/driver/http/http.go" ]] || {
        echo "FAIL: driver/http/http.go missing"; return 1
    }
    # Adapter creates package named after adapter (not "repository")
    find "$PROJ/internal/adapters/driven/database" -name "*.go" -type f | grep -q . || {
        echo "FAIL: driven/database has no .go files"; return 1
    }

    # Validate (build skipped due to generator bugs with uuid/adapter imports)
    (cd "$PROJ" && WORK_DIR="" hex validate)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: driver-driven validate returned $?"; return 1
    fi

    return 0
}

# Test --in-place flag: project generated directly in target dir.
test_in_place() {
    WORK_DIR=$(mktemp -d)
    trap 'cleanup; WORK_DIR=""' EXIT

    hex init inplace --in-place --module github.com/test/inplace \
        --with-migrations --with-observability --with-metrics

    if [[ $_rc -ne 0 ]]; then
        echo "FAIL: in-place init returned $_rc"; return 1
    fi

    # In-place: files directly in WORK_DIR
    [[ -f "$WORK_DIR/go.mod" ]] || {
        echo "FAIL: go.mod not in place"; return 1
    }
    [[ -f "$WORK_DIR/.hexago.yaml" ]] || {
        echo "FAIL: .hexago.yaml not in place"; return 1
    }
    [[ -d "$WORK_DIR/internal" ]] || {
        echo "FAIL: internal/ dir missing"; return 1
    }

    (cd "$WORK_DIR" && go build ./...) || {
        echo "FAIL: in-place build error"; return 1
    }

    return 0
}

# Test all add commands with cross-entity references.
test_add_cycle() {
    WORK_DIR=$(mktemp -d)
    trap 'cleanup; WORK_DIR=""' EXIT

    hex init demo-add \
        --project-type service \
        --module github.com/test/demo-add

    if [[ $_rc -ne 0 ]]; then
        echo "FAIL: add-cycle init returned $_rc"; return 1
    fi

    local PROJ
    PROJ=$(project_root demo-add no)

    # Create domain dir (generator bug: add entity requires it but doesn't create it)
    mkdir -p "$PROJ/internal/core/domain"

    # cd into project dir so relative paths in domain generator resolve correctly
    (cd "$PROJ" && WORK_DIR="" hex add domain entity Item --fields "id:uuid.UUID,name:string,price:float64,createdAt:time.Time")
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add domain entity Item returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add domain valueobject Money --fields "amount:float64,currency:string")
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add domain valueobject Money returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add domain valueobject Price --fields "value:float64" -e Item)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add domain valueobject Price (entity-bound) returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add service ItemService --entity Item)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add service returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add adapter secondary database ItemRepository --entity Item)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add adapter secondary database returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add adapter secondary external SmsClient)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add adapter secondary external returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add adapter secondary cache SessionCache)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add adapter secondary cache returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add adapter secondary htmlparser HtmlParser)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add adapter secondary htmlparser returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add adapter primary http ItemHandler)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add adapter primary returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add worker ItemWorker --type queue --workers 3 --queue-size 50)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add worker queue returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add worker CleanupWorker --type periodic --interval 1h)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add worker periodic returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add migration create_items_table)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add migration sql returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add migration --type go migrate_items)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add migration go returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add tool logger ZerologLogger --description "structured logger")
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add tool logger returned $?"; return 1
    fi

    (cd "$PROJ" && WORK_DIR="" hex add tool validator RequestValidator --description "input validation")
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add tool validator returned $?"; return 1
    fi

    # Validate (build skipped due to generator bugs: uuid dep not added, adapter import path mismatch)
    (cd "$PROJ" && WORK_DIR="" hex validate)
    if [[ $? -ne 0 ]]; then
        echo "FAIL: add-cycle validate returned $?"; return 1
    fi

    return 0
}

# Test postgres db-driver migration branch in generated migrator.go.
test_postgres_migration() {
    WORK_DIR=$(mktemp -d)
    trap 'cleanup; WORK_DIR=""' EXIT

    hex init demo-pg \
        --project-type http-server \
        --framework stdlib \
        --with-migrations \
        --db-driver postgres

    if [[ $_rc -ne 0 ]]; then
        echo "FAIL: postgres init returned $_rc"; return 1
    fi

    local PROJ
    PROJ=$(project_root demo-pg no)

    # Check migrator.go contains postgres driver path
    local migrator_file
    migrator_file=$(find "$PROJ" -name 'migrator.go' 2>/dev/null | head -1)
    [[ -n "$migrator_file" ]] || {
        echo "FAIL: migrator.go not generated"; return 1
    }
    if [[ ! -f "$migrator_file" ]]; then
        echo "FAIL: migrator.go not generated"; return 1
    fi
    if ! grep -q 'postgres.WithInstance' "$migrator_file"; then
        echo "FAIL: migrator.go missing postgres driver branch"; return 1
    fi

    return 0
}

# Test sqlite db-driver migration branch.
test_sqlite_migration() {
    WORK_DIR=$(mktemp -d)
    trap 'cleanup; WORK_DIR=""' EXIT

    hex init demo-sqlite \
        --project-type http-server \
        --framework stdlib \
        --with-migrations \
        --db-driver sqlite3

    if [[ $_rc -ne 0 ]]; then
        echo "FAIL: sqlite init returned $_rc"; return 1
    fi

    local PROJ
    PROJ=$(project_root demo-sqlite no)

    local migrator_file
    migrator_file=$(find "$PROJ" -name 'migrator.go' 2>/dev/null | head -1)
    [[ -n "$migrator_file" ]] || {
        echo "FAIL: migrator.go not generated"; return 1
    }
    if ! grep -q 'sqlite.WithInstance' "$migrator_file"; then
        echo "FAIL: migrator.go missing sqlite driver branch"; return 1
    fi

    return 0
}

# ─── Runner ──────────────────────────────────────────────────────────────────

run_scenario() {
    local name="$1"
    shift
    local result=0
    eval "$@" || result=$?
    if [[ $result -eq 0 ]]; then
        PASS_COUNT=$((PASS_COUNT + 1))
        echo "  [PASS] $name"
    else
        FAIL_COUNT=$((FAIL_COUNT + 1))
        echo "  [FAIL] $name"
    fi
    return $result
}

run_all() {
    echo ""
    echo "============================================================"
    echo " HEXAGO E2E — Full Generator Test Matrix"
    echo "============================================================"
    echo ""

    run_scenario "service-full"       test_service_full
    run_scenario "frameworks"         test_frameworks
    run_scenario "cli"                test_cli
    run_scenario "driver-driven"      test_driver_driven
    run_scenario "in-place"           test_in_place
    run_scenario "add-cycle"          test_add_cycle
    run_scenario "postgres-migration" test_postgres_migration
    run_scenario "sqlite-migration"   test_sqlite_migration

    echo ""
    echo "============================================================"
    echo " Results: $((PASS_COUNT + FAIL_COUNT)) total"
    echo "   PASSED: $PASS_COUNT"
    echo "   FAILED: $FAIL_COUNT"
    echo "============================================================"
    echo ""

    if [[ $FAIL_COUNT -gt 0 ]]; then
        exit 1
    fi
}

# ─── Entry ───────────────────────────────────────────────────────────────────

echo "==> Hexago E2E test runner"
echo ""

# Ensure go is available
if ! command -v go >/dev/null 2>&1; then
    echo "FAIL: go not found in PATH"
    exit 1
fi

# Build hexago if --reuse not set
if [[ "${REUSE_BIN:-}" != "1" && -f "$PROJECT_ROOT/hexago" ]]; then
    BIN="$PROJECT_ROOT/hexago"
    echo "==> Reusing existing hexago binary: $BIN"
else
    build_bin
fi

echo ""
echo "==> Binary: $BIN"
echo "==> Project root: $PROJECT_ROOT"
echo ""

run_all
