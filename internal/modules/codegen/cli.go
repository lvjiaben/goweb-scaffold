package codegen

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/lvjiaben/goweb-scaffold/internal/gen/service"
)

type CLIBackend interface {
	Tables() ([]BusinessTable, error)
	Modules() ([]service.ManagedModule, error)
	Preview(moduleName string, tableName string, payload json.RawMessage) (service.Preview, []service.ColumnInfo, error)
	Diff(input ActionInput) (service.DiffResult, error)
	Generate(input ActionInput) (service.GenerateResult, error)
	Regenerate(input RegenerateInput) (service.GenerateResult, error)
	Remove(input service.RemoveInput) (service.RemoveResult, error)
	Export(input ExportInput) (service.ExportFile, error)
	Import(input ImportInput) (ImportResult, error)
	ResolveInput(input SourceInput) (ResolvedInput, error)
}

type CLI struct {
	backend CLIBackend
	stdout  io.Writer
	stderr  io.Writer
}

func NewCLI(backend CLIBackend, stdout io.Writer, stderr io.Writer) *CLI {
	return &CLI{
		backend: backend,
		stdout:  stdout,
		stderr:  stderr,
	}
}

func DetectConfigPath(args []string) string {
	return detectFlagValue(args, "-config", "configs/config.yaml")
}

func DetectFormat(args []string) string {
	return detectFlagValue(args, "-format", "text")
}

func detectFlagValue(args []string, flagName string, fallback string) string {
	for index := 0; index < len(args); index++ {
		arg := strings.TrimSpace(args[index])
		switch {
		case strings.HasPrefix(arg, flagName+"="):
			value := strings.TrimSpace(strings.TrimPrefix(arg, flagName+"="))
			if value != "" {
				return value
			}
		case arg == flagName && index+1 < len(args):
			value := strings.TrimSpace(args[index+1])
			if value != "" {
				return value
			}
		}
	}
	return fallback
}

func (c *CLI) Run(args []string) int {
	if len(args) == 0 {
		c.writeUsage()
		return 1
	}

	switch args[0] {
	case "tables":
		return c.runTables(args[1:])
	case "modules":
		return c.runModules(args[1:])
	case "preview":
		return c.runPreview(args[1:])
	case "diff":
		return c.runDiff(args[1:])
	case "generate":
		return c.runGenerate(args[1:])
	case "regenerate":
		return c.runRegenerate(args[1:])
	case "remove":
		return c.runRemove(args[1:])
	case "export":
		return c.runExport(args[1:])
	case "import":
		return c.runImport(args[1:])
	case "-h", "--help", "help":
		c.writeUsage()
		return 0
	default:
		c.writeTextError(fmt.Errorf("unknown subcommand: %s", args[0]))
		c.writeUsage()
		return 1
	}
}

type commonFlags struct {
	configPath string
	format     string
	outputPath string
}

type previewCommand struct {
	commonFlags
	moduleName string
	tableName  string
	payload    string
	from       string
}

type actionCommand struct {
	commonFlags
	moduleName     string
	tableName      string
	payload        string
	from           string
	overwrite      bool
	registerModule bool
	upsertMenu     bool
}

type regenerateCommand struct {
	commonFlags
	moduleName     string
	historyID      int64
	overwrite      bool
	registerModule bool
	upsertMenu     bool
}

type removeCommand struct {
	commonFlags
	moduleName       string
	removeFiles      bool
	unregisterModule bool
	removeMenu       bool
	removeHistory    bool
	removeLock       bool
}

type exportCommand struct {
	commonFlags
	moduleName string
	historyID  int64
}

type importCommand struct {
	commonFlags
	from           string
	moduleName     string
	tableName      string
	payload        string
	diff           bool
	generate       bool
	dryRun         bool
	overwrite      bool
	registerModule bool
	upsertMenu     bool
}

func parsePreviewCommand(args []string) (previewCommand, error) {
	cmd := previewCommand{}
	fs := newFlagSet("preview")
	registerCommonFlags(fs, &cmd.commonFlags)
	fs.StringVar(&cmd.moduleName, "module", "", "module name")
	fs.StringVar(&cmd.tableName, "table", "", "table name")
	fs.StringVar(&cmd.payload, "payload", "", "payload json file")
	fs.StringVar(&cmd.from, "from", "", "load from export or lock file")
	if err := fs.Parse(args); err != nil {
		return cmd, err
	}
	return cmd, nil
}

func parseActionCommand(name string, args []string) (actionCommand, error) {
	cmd := actionCommand{
		overwrite:      true,
		registerModule: true,
		upsertMenu:     true,
	}
	fs := newFlagSet(name)
	registerCommonFlags(fs, &cmd.commonFlags)
	fs.StringVar(&cmd.moduleName, "module", "", "module name")
	fs.StringVar(&cmd.tableName, "table", "", "table name")
	fs.StringVar(&cmd.payload, "payload", "", "payload json file")
	fs.StringVar(&cmd.from, "from", "", "load from export or lock file")
	fs.BoolVar(&cmd.overwrite, "overwrite", true, "overwrite generator-managed files")
	fs.BoolVar(&cmd.registerModule, "register-module", true, "rebuild generated registry files")
	fs.BoolVar(&cmd.upsertMenu, "upsert-menu", true, "upsert admin menu and role-menu")
	if err := fs.Parse(args); err != nil {
		return cmd, err
	}
	return cmd, nil
}

func parseRegenerateCommand(args []string) (regenerateCommand, error) {
	cmd := regenerateCommand{
		overwrite:      true,
		registerModule: true,
		upsertMenu:     true,
	}
	fs := newFlagSet("regenerate")
	registerCommonFlags(fs, &cmd.commonFlags)
	fs.StringVar(&cmd.moduleName, "module", "", "module name")
	fs.Int64Var(&cmd.historyID, "history-id", 0, "history id")
	fs.BoolVar(&cmd.overwrite, "overwrite", true, "overwrite generator-managed files")
	fs.BoolVar(&cmd.registerModule, "register-module", true, "rebuild generated registry files")
	fs.BoolVar(&cmd.upsertMenu, "upsert-menu", true, "upsert admin menu and role-menu")
	if err := fs.Parse(args); err != nil {
		return cmd, err
	}
	return cmd, nil
}

func parseRemoveCommand(args []string) (removeCommand, error) {
	cmd := removeCommand{
		removeFiles:      true,
		unregisterModule: true,
		removeMenu:       true,
		removeLock:       true,
	}
	fs := newFlagSet("remove")
	registerCommonFlags(fs, &cmd.commonFlags)
	fs.StringVar(&cmd.moduleName, "module", "", "module name")
	fs.BoolVar(&cmd.removeFiles, "remove-files", true, "remove generated files")
	fs.BoolVar(&cmd.unregisterModule, "unregister-module", true, "rebuild generated registry files without this module")
	fs.BoolVar(&cmd.removeMenu, "remove-menu", true, "remove module menu and role-menu links")
	fs.BoolVar(&cmd.removeHistory, "remove-history", false, "remove module history")
	fs.BoolVar(&cmd.removeLock, "remove-lock", true, "remove codegen lock file")
	if err := fs.Parse(args); err != nil {
		return cmd, err
	}
	return cmd, nil
}

func parseExportCommand(args []string) (exportCommand, error) {
	cmd := exportCommand{}
	fs := newFlagSet("export")
	registerCommonFlags(fs, &cmd.commonFlags)
	fs.StringVar(&cmd.moduleName, "module", "", "module name")
	fs.Int64Var(&cmd.historyID, "history-id", 0, "history id")
	if err := fs.Parse(args); err != nil {
		return cmd, err
	}
	return cmd, nil
}

func parseImportCommand(args []string) (importCommand, error) {
	cmd := importCommand{
		overwrite:      true,
		registerModule: true,
		upsertMenu:     true,
	}
	fs := newFlagSet("import")
	registerCommonFlags(fs, &cmd.commonFlags)
	fs.StringVar(&cmd.from, "from", "", "import from export or lock file")
	fs.StringVar(&cmd.moduleName, "module", "", "override module name")
	fs.StringVar(&cmd.tableName, "table", "", "override table name")
	fs.StringVar(&cmd.payload, "payload", "", "override payload json file")
	fs.BoolVar(&cmd.diff, "diff", false, "run diff instead of preview")
	fs.BoolVar(&cmd.generate, "generate", false, "run generate")
	fs.BoolVar(&cmd.dryRun, "dry-run", false, "alias of --diff")
	fs.BoolVar(&cmd.overwrite, "overwrite", true, "overwrite generator-managed files")
	fs.BoolVar(&cmd.registerModule, "register-module", true, "rebuild generated registry files")
	fs.BoolVar(&cmd.upsertMenu, "upsert-menu", true, "upsert admin menu and role-menu")
	if err := fs.Parse(args); err != nil {
		return cmd, err
	}
	return cmd, nil
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func registerCommonFlags(fs *flag.FlagSet, common *commonFlags) {
	fs.StringVar(&common.configPath, "config", "configs/config.yaml", "config file path")
	fs.StringVar(&common.format, "format", "text", "output format: text or json")
	fs.StringVar(&common.outputPath, "output", "", "write JSON payload to file")
}

func (c *CLI) runTables(args []string) int {
	common, err := parseSimpleCommand("tables", args)
	if err != nil {
		return c.fail("text", err)
	}
	rows, err := c.backend.Tables()
	if err != nil {
		return c.fail(common.format, err)
	}
	if common.format == "json" {
		return c.writeJSON(map[string]any{"list": rows}, "")
	}

	var buffer bytes.Buffer
	writer := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(writer, "TABLE_NAME\tDISPLAY_NAME\tTABLE_COMMENT")
	for _, row := range rows {
		_, _ = fmt.Fprintf(writer, "%s\t%s\t%s\n", row.TableName, row.DisplayName, row.TableComment)
	}
	_ = writer.Flush()
	_, _ = io.WriteString(c.stdout, buffer.String())
	return 0
}

func (c *CLI) runModules(args []string) int {
	common, err := parseSimpleCommand("modules", args)
	if err != nil {
		return c.fail("text", err)
	}
	rows, err := c.backend.Modules()
	if err != nil {
		return c.fail(common.format, err)
	}
	if common.format == "json" {
		return c.writeJSON(map[string]any{"list": rows}, "")
	}

	var buffer bytes.Buffer
	writer := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(writer, "MODULE_NAME\tTABLE_NAME\tGENERATED_AT\tTEMPLATE_VERSION\tROUTE_PATH")
	for _, row := range rows {
		_, _ = fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n", row.ModuleName, row.TableName, row.GeneratedAt, row.TemplateVersion, row.RoutePath)
	}
	_ = writer.Flush()
	_, _ = io.WriteString(c.stdout, buffer.String())
	return 0
}

func (c *CLI) runPreview(args []string) int {
	cmd, err := parsePreviewCommand(args)
	if err != nil {
		return c.fail("text", err)
	}

	resolved, err := c.backend.ResolveInput(SourceInput{
		ModuleName:  cmd.moduleName,
		TableName:   cmd.tableName,
		PayloadPath: cmd.payload,
		FromPath:    cmd.from,
	})
	if err != nil {
		return c.fail(cmd.format, err)
	}

	previewPayload, _, err := c.backend.Preview(resolved.ModuleName, resolved.TableName, resolved.Payload)
	if err != nil {
		return c.fail(cmd.format, err)
	}
	return c.writePreview(cmd.commonFlags, resolved, previewPayload)
}

func (c *CLI) runDiff(args []string) int {
	cmd, err := parseActionCommand("diff", args)
	if err != nil {
		return c.fail("text", err)
	}
	resolved, err := c.backend.ResolveInput(SourceInput{
		ModuleName:  cmd.moduleName,
		TableName:   cmd.tableName,
		PayloadPath: cmd.payload,
		FromPath:    cmd.from,
	})
	if err != nil {
		return c.fail(cmd.format, err)
	}
	result, err := c.backend.Diff(ActionInput{
		ModuleName:     resolved.ModuleName,
		TableName:      resolved.TableName,
		Payload:        resolved.Payload,
		Overwrite:      cmd.overwrite,
		RegisterModule: cmd.registerModule,
		UpsertMenu:     cmd.upsertMenu,
	})
	if err != nil {
		return c.fail(cmd.format, err)
	}
	if cmd.format == "json" {
		return c.writeJSON(result, cmd.outputPath)
	}
	c.writeTextSection("diff", result)
	return 0
}

func (c *CLI) runGenerate(args []string) int {
	cmd, err := parseActionCommand("generate", args)
	if err != nil {
		return c.fail("text", err)
	}
	resolved, err := c.backend.ResolveInput(SourceInput{
		ModuleName:  cmd.moduleName,
		TableName:   cmd.tableName,
		PayloadPath: cmd.payload,
		FromPath:    cmd.from,
	})
	if err != nil {
		return c.fail(cmd.format, err)
	}
	result, err := c.backend.Generate(ActionInput{
		ModuleName:     resolved.ModuleName,
		TableName:      resolved.TableName,
		Payload:        resolved.Payload,
		Overwrite:      cmd.overwrite,
		RegisterModule: cmd.registerModule,
		UpsertMenu:     cmd.upsertMenu,
	})
	if err != nil {
		return c.fail(cmd.format, err)
	}
	if cmd.format == "json" {
		return c.writeJSON(result, cmd.outputPath)
	}
	c.writeTextSection("generate", result)
	return 0
}

func (c *CLI) runRegenerate(args []string) int {
	cmd, err := parseRegenerateCommand(args)
	if err != nil {
		return c.fail("text", err)
	}
	if strings.TrimSpace(cmd.moduleName) == "" && cmd.historyID <= 0 {
		return c.fail(cmd.format, errors.New("module or history-id is required"))
	}
	result, err := c.backend.Regenerate(RegenerateInput{
		ModuleName:     cmd.moduleName,
		HistoryID:      cmd.historyID,
		Overwrite:      cmd.overwrite,
		RegisterModule: cmd.registerModule,
		UpsertMenu:     cmd.upsertMenu,
	})
	if err != nil {
		return c.fail(cmd.format, err)
	}
	if cmd.format == "json" {
		return c.writeJSON(result, cmd.outputPath)
	}
	c.writeTextSection("regenerate", result)
	return 0
}

func (c *CLI) runRemove(args []string) int {
	cmd, err := parseRemoveCommand(args)
	if err != nil {
		return c.fail("text", err)
	}
	if strings.TrimSpace(cmd.moduleName) == "" {
		return c.fail(cmd.format, errors.New("module is required"))
	}
	result, err := c.backend.Remove(service.RemoveInput{
		ModuleName:       cmd.moduleName,
		RemoveFiles:      cmd.removeFiles,
		UnregisterModule: cmd.unregisterModule,
		RemoveMenu:       cmd.removeMenu,
		RemoveHistory:    cmd.removeHistory,
		RemoveLock:       cmd.removeLock,
	})
	if err != nil {
		return c.fail(cmd.format, err)
	}
	if cmd.format == "json" {
		return c.writeJSON(result, cmd.outputPath)
	}
	c.writeTextSection("remove", result)
	return 0
}

func (c *CLI) runExport(args []string) int {
	cmd, err := parseExportCommand(args)
	if err != nil {
		return c.fail("text", err)
	}
	if strings.TrimSpace(cmd.moduleName) == "" && cmd.historyID <= 0 {
		return c.fail(cmd.format, errors.New("module or history-id is required"))
	}
	result, err := c.backend.Export(ExportInput{
		ModuleName: cmd.moduleName,
		HistoryID:  cmd.historyID,
	})
	if err != nil {
		return c.fail(cmd.format, err)
	}
	if cmd.format == "json" {
		return c.writeJSON(result, cmd.outputPath)
	}
	if cmd.outputPath != "" {
		raw, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return c.fail(cmd.format, err)
		}
		if err := os.WriteFile(cmd.outputPath, append(raw, '\n'), 0o644); err != nil {
			return c.fail(cmd.format, err)
		}
		_, _ = fmt.Fprintf(c.stdout, "export written to %s\n", cmd.outputPath)
		return 0
	}
	c.writeTextSection("export", result)
	return 0
}

func (c *CLI) runImport(args []string) int {
	cmd, err := parseImportCommand(args)
	if err != nil {
		return c.fail("text", err)
	}
	if strings.TrimSpace(cmd.from) == "" {
		return c.fail(cmd.format, errors.New("from is required"))
	}
	if cmd.generate && cmd.dryRun {
		return c.fail(cmd.format, errors.New("generate and dry-run cannot be used together"))
	}

	mode := ImportModePreview
	if cmd.generate {
		mode = ImportModeGenerate
	} else if cmd.diff || cmd.dryRun {
		mode = ImportModeDiff
	}

	result, err := c.backend.Import(ImportInput{
		FromPath:       cmd.from,
		ModuleName:     cmd.moduleName,
		TableName:      cmd.tableName,
		PayloadPath:    cmd.payload,
		Mode:           mode,
		Overwrite:      cmd.overwrite,
		RegisterModule: cmd.registerModule,
		UpsertMenu:     cmd.upsertMenu,
	})
	if err != nil {
		return c.fail(cmd.format, err)
	}
	if cmd.format == "json" {
		return c.writeJSON(result, cmd.outputPath)
	}
	c.writeTextSection("import", result)
	return 0
}

func parseSimpleCommand(name string, args []string) (commonFlags, error) {
	var common commonFlags
	fs := newFlagSet(name)
	registerCommonFlags(fs, &common)
	if err := fs.Parse(args); err != nil {
		return common, err
	}
	return common, nil
}

func (c *CLI) writePreview(common commonFlags, resolved ResolvedInput, previewPayload service.Preview) int {
	if common.format == "json" {
		payload := map[string]any{
			"source_kind": resolved.SourceKind,
			"preview":     previewPayload,
		}
		return c.writeJSON(payload, common.outputPath)
	}

	if common.outputPath != "" {
		if code := c.writeJSON(previewPayload, common.outputPath); code != 0 {
			return code
		}
		_, _ = fmt.Fprintf(c.stdout, "preview written to %s\n", common.outputPath)
		return 0
	}

	_, _ = fmt.Fprintf(c.stdout, "source: %s\nmodule: %s\ntable: %s\nroute: %s\napi: %s\nnotes: %d\n",
		firstNonEmptyString(resolved.SourceKind, "direct"),
		previewPayload.ModuleName,
		previewPayload.TableName,
		previewPayload.Page.RoutePath,
		previewPayload.API.ModuleCode,
		len(previewPayload.Notes),
	)
	return 0
}

func (c *CLI) writeJSON(value any, outputPath string) int {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		c.writeTextError(err)
		return 1
	}
	if strings.TrimSpace(outputPath) != "" {
		if err := os.WriteFile(outputPath, append(raw, '\n'), 0o644); err != nil {
			c.writeTextError(err)
			return 1
		}
	}
	_, _ = c.stdout.Write(append(raw, '\n'))
	return 0
}

func (c *CLI) fail(format string, err error) int {
	if strings.TrimSpace(format) == "json" {
		_ = c.writeJSON(map[string]any{"error": err.Error()}, "")
		return 1
	}
	c.writeTextError(err)
	return 1
}

func (c *CLI) writeTextError(err error) {
	_, _ = fmt.Fprintf(c.stderr, "error: %s\n", err.Error())
}

func (c *CLI) writeUsage() {
	_, _ = io.WriteString(c.stderr, `Usage: codegen <subcommand> [flags]

Subcommands:
  tables
  modules
  preview
  diff
  generate
  regenerate
  remove
  export
  import
`)
}

func (c *CLI) writeTextSection(title string, value any) {
	_, _ = fmt.Fprintf(c.stdout, "[%s]\n", title)
	pretty, _ := json.MarshalIndent(value, "", "  ")
	_, _ = c.stdout.Write(append(pretty, '\n'))
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
