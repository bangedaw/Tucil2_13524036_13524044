package main

import (
	"bufio"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)



// Model menyimpan data geometri dari berkas .obj.
type Model struct {
	Vertices []Vector3
	Faces    [][]int
}

var currentModel Model

// ObjtoModel membaca dan memuat berkas .obj ke dalam struktur Model.
func ObjtoModel(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var vertices []Vector3
	var faces [][]int
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "v": // Titik Sudut (Vertex)
			if len(parts) >= 4 {
				x, _ := strconv.ParseFloat(parts[1], 64)
				y, _ := strconv.ParseFloat(parts[2], 64)
				z, _ := strconv.ParseFloat(parts[3], 64)
				vertices = append(vertices, Vector3{X: x, Y: y, Z: z})
			}
		case "f": // Permukaan (Face)
			var face []int
			for _, p := range parts[1:] {
				// Format .obj bisa v, v/vt, atau v/vt/vn. Kita hanya mengambil v.
				vStr := strings.Split(p, "/")[0]
				idx, err := strconv.Atoi(vStr)
				if err == nil {
					// Indeks .obj dimulai dari 1, sedangkan array Go dari 0.
					if idx > 0 {
						face = append(face, idx-1)
					} else {
						// Menangani indeks negatif (relatif)
						face = append(face, len(vertices)+idx)
					}
				}
			}
			if len(face) >= 3 {
				faces = append(faces, face)
			}
		}
	}

	currentModel = normalizeModel(Model{Vertices: vertices, Faces: faces})
	return scanner.Err()
}

// normalizeModel menempatkan model di pusat origin (0,0,0) dan menyesuaikan ukurannya.
func normalizeModel(m Model) Model {
	if len(m.Vertices) == 0 {
		return m
	}

	minV, maxV := m.Vertices[0], m.Vertices[0]
	for _, v := range m.Vertices {
		minV.X, maxV.X = math.Min(minV.X, v.X), math.Max(maxV.X, v.X)
		minV.Y, maxV.Y = math.Min(minV.Y, v.Y), math.Max(maxV.Y, v.Y)
		minV.Z, maxV.Z = math.Min(minV.Z, v.Z), math.Max(maxV.Z, v.Z)
	}

	centerX := (minV.X + maxV.X) / 2
	centerY := (minV.Y + maxV.Y) / 2
	centerZ := (minV.Z + maxV.Z) / 2

	maxDim := math.Max(maxV.X-minV.X, math.Max(maxV.Y-minV.Y, maxV.Z-minV.Z))
	scale := 2.0 / maxDim

	for i := range m.Vertices {
		m.Vertices[i].X = (m.Vertices[i].X - centerX) * scale
		m.Vertices[i].Y = (m.Vertices[i].Y - centerY) * scale
		m.Vertices[i].Z = (m.Vertices[i].Z - centerZ) * scale
	}

	return m
}

// renderSVG memproyeksikan titik 3D ke 2D dan menghasilkan format SVG.
func renderSVG(m Model, rx, ry, scale float64) string {
	var sb strings.Builder
	width, height := 800.0, 600.0
	centerX, centerY := width/2, height/2

	sb.WriteString(fmt.Sprintf(`<svg width="100%%" height="100%%" viewBox="0 0 %.0f %.0f" xmlns="http://www.w3.org/2000/svg" style="background-color: #f4f4f9;">`, width, height))

	// Konversi sudut rotasi ke radian
	sinX, cosX := math.Sin(rx), math.Cos(rx)
	sinY, cosY := math.Sin(ry), math.Cos(ry)

	projected := make([]Vector3, len(m.Vertices))
	for i, v := range m.Vertices {
		// Rotasi sumbu Y
		x1 := v.X*cosY - v.Z*sinY
		z1 := v.X*sinY + v.Z*cosY
		y1 := v.Y

		// Rotasi sumbu X
		y2 := y1*cosX - z1*sinX
		// z2 := y1*sinX + z1*cosX (Z tidak diperlukan untuk proyeksi ortografis 2D)

		// Proyeksi ortografis ke bidang 2D dan perbesaran skala
		// Sumbu Y dibalik karena koordinat SVG dimulai dari kiri atas
		projected[i] = Vector3{
			X: x1*scale + centerX,
			Y: -y2*scale + centerY, 
		}
	}

	// Menggambar struktur kawat (wireframe)
	sb.WriteString(`<g stroke="#2c3e50" stroke-width="1.5" fill="none">`)
	for _, face := range m.Faces {
		if len(face) < 3 {
			continue
		}
		sb.WriteString(`<polygon points="`)
		for _, idx := range face {
			if idx >= 0 && idx < len(projected) {
				sb.WriteString(fmt.Sprintf("%.2f,%.2f ", projected[idx].X, projected[idx].Y))
			}
		}
		sb.WriteString(`" />`)
	}
	sb.WriteString(`</g></svg>`)

	return sb.String()
}

// handleIndex melayani halaman web antarmuka pengguna.
func handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <title>Go 3D Object Viewer</title>
    <style>
        body { margin: 0; overflow: hidden; display: flex; flex-direction: column; height: 100vh; font-family: sans-serif; }
        #viewer { flex-grow: 1; cursor: grab; }
        #viewer:active { cursor: grabbing; }
        .controls { position: absolute; top: 10px; left: 10px; background: rgba(255,255,255,0.8); padding: 10px; border-radius: 5px; box-shadow: 0 2px 5px rgba(0,0,0,0.2); }
    </style>
</head>
<body>
    <div class="controls">
        <h3>Go 3D Viewer</h3>
        <p>Tahan klik dan geser untuk <b>Rotasi</b>.<br>Gunakan scroll wheel untuk <b>Zoom</b>.</p>
    </div>
    <div id="viewer"></div>

    <script>
        let rx = 0;      // Rotasi X dalam Radian
        let ry = 0;      // Rotasi Y dalam Radian
        let scale = 200; // Skala Perbesaran Dasar
        
        let isDragging = false;
        let lastMouseX = 0;
        let lastMouseY = 0;
        let isRendering = false;

        const viewer = document.getElementById('viewer');

        function requestRender() {
            if (isRendering) return;
            isRendering = true;
            fetch('/render?rx=' + rx + '&ry=' + ry + '&scale=' + scale)
                .then(response => response.text())
                .then(svg => {
                    viewer.innerHTML = svg;
                    isRendering = false;
                });
        }

        viewer.addEventListener('mousedown', (e) => {
            isDragging = true;
            lastMouseX = e.clientX;
            lastMouseY = e.clientY;
        });

        window.addEventListener('mouseup', () => {
            isDragging = false;
        });

        window.addEventListener('mousemove', (e) => {
            if (!isDragging) return;
            const deltaX = e.clientX - lastMouseX;
            const deltaY = e.clientY - lastMouseY;
            
            // Sensitivitas Rotasi
            ry += deltaX * 0.01;
            rx += deltaY * 0.01;
            
            lastMouseX = e.clientX;
            lastMouseY = e.clientY;
            
            requestRender();
        });

        viewer.addEventListener('wheel', (e) => {
            e.preventDefault();
            scale += e.deltaY * -0.2;
            if (scale < 10) scale = 10; 
            requestRender();
        }, { passive: false });

        requestRender();
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleRender menerima parameter transformasi dari klien dan mengembalikan SVG.
func handleRender(w http.ResponseWriter, r *http.Request) {
	rx, _ := strconv.ParseFloat(r.URL.Query().Get("rx"), 64)
	ry, _ := strconv.ParseFloat(r.URL.Query().Get("ry"), 64)
	scale, _ := strconv.ParseFloat(r.URL.Query().Get("scale"), 64)

	if scale == 0 {
		scale = 200 // Skala bawaan (fallback)
	}

	svgOutput := renderSVG(currentModel, rx, ry, scale)
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write([]byte(svgOutput))
}
