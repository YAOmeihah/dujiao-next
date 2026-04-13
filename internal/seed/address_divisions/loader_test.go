package addressdivisions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDatasetFromDir(t *testing.T) {
	dir := t.TempDir()

	writeDatasetFile(t, dir, "provinces.json", `[{"code":"33","name":"浙江省"}]`)
	writeDatasetFile(t, dir, "cities.json", `[{"code":"3301","name":"杭州市","provinceCode":"33"}]`)
	writeDatasetFile(t, dir, "districts.json", `[{"code":"330106","name":"西湖区","cityCode":"3301","provinceCode":"33"}]`)
	writeDatasetFile(t, dir, "townships.json", `[{"code":"330106001","name":"西溪街道","areaCode":"330106","cityCode":"3301","provinceCode":"33"}]`)
	writeDatasetFile(t, dir, "villages.json", `[{"code":"330106001001","name":"文一社区","streetCode":"330106001","areaCode":"330106","cityCode":"3301","provinceCode":"33"}]`)

	dataset, err := LoadDatasetFromDir(dir)
	if err != nil {
		t.Fatalf("LoadDatasetFromDir failed: %v", err)
	}

	if len(dataset.Provinces) != 1 || dataset.Provinces[0].Name != "浙江省" {
		t.Fatalf("unexpected provinces: %+v", dataset.Provinces)
	}
	if len(dataset.Villages) != 1 || dataset.Villages[0].TownshipCode != "330106001" {
		t.Fatalf("unexpected villages: %+v", dataset.Villages)
	}
}

func TestLoadDatasetFromDirRequiresFiles(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadDatasetFromDir(dir)
	if err == nil {
		t.Fatal("expected LoadDatasetFromDir to fail when files are missing")
	}
}

func writeDatasetFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s failed: %v", name, err)
	}
}
