package main

import (
	"bufio"
	"encoding/json"
	"html/template"
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
// (Fungsi ini tetap sama persis seperti buatan temanmu)
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
		case "v":
			if len(parts) >= 4 {
				x, _ := strconv.ParseFloat(parts[1], 64)
				y, _ := strconv.ParseFloat(parts[2], 64)
				z, _ := strconv.ParseFloat(parts[3], 64)
				vertices = append(vertices, Vector3{X: x, Y: y, Z: z})
			}
		case "f":
			var face []int
			for _, p := range parts[1:] {
				vStr := strings.Split(p, "/")[0]
				idx, err := strconv.Atoi(vStr)
				if err == nil {
					if idx > 0 {
						face = append(face, idx-1)
					} else {
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
// (Fungsi ini juga tetap sama)
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

// handleIndex menyajikan halaman web tunggal sekaligus menyuntikkan data JSON
func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Konversi struktur data Model Go kita menjadi format JSON
	modelJSON, err := json.Marshal(currentModel)
	if err != nil {
		http.Error(w, "Gagal memproses model", http.StatusInternalServerError)
		return
	}

	// HTML & JavaScript Terintegrasi
	html := `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <title>Go 3D Object Viewer - Optimized</title>
    <style>
        body { margin: 0; overflow: hidden; background-color: #f4f4f9; font-family: sans-serif; }
        canvas { display: block; cursor: grab; }
        canvas:active { cursor: grabbing; }
        .controls { position: absolute; top: 10px; left: 10px; background: rgba(255,255,255,0.9); padding: 15px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
    </style>
</head>
<body>
    <div class="controls">
        <h3 style="margin-top:0;">Go 3D Viewer (Client-Side)</h3>
        <p style="margin-bottom:0;">Tahan klik dan geser untuk <b>Rotasi</b>.<br>Gunakan scroll wheel untuk <b>Zoom</b>.</p>
    </div>
    
    <canvas id="viewer"></canvas>

    <script>
        // Menerima suntikan data model 3D (Vertices dan Faces) dari server Go
        const model = {{.ModelData}};

        const canvas = document.getElementById('viewer');
        const ctx = canvas.getContext('2d');

        let width = window.innerWidth;
        let height = window.innerHeight;
        canvas.width = width;
        canvas.height = height;

        // Auto-resize canvas jika jendela browser diubah ukurannya
        window.addEventListener('resize', () => {
            width = window.innerWidth;
            height = window.innerHeight;
            canvas.width = width;
            canvas.height = height;
            draw();
        });

        let rx = 0.5; // Rotasi awal X
        let ry = -0.5; // Rotasi awal Y
        let scale = Math.min(width, height) / 3;

        let isDragging = false;
        let lastX = 0;
        let lastY = 0;

        canvas.addEventListener('mousedown', (e) => {
            isDragging = true;
            lastX = e.clientX;
            lastY = e.clientY;
        });

        window.addEventListener('mouseup', () => {
            isDragging = false;
        });

        window.addEventListener('mousemove', (e) => {
            if (!isDragging) return;
            const deltaX = e.clientX - lastX;
            const deltaY = e.clientY - lastY;
            
            ry += deltaX * 0.01;
            rx += deltaY * 0.01;
            
            lastX = e.clientX;
            lastY = e.clientY;
            
            // requestAnimationFrame memastikan render mulus sesuai refresh rate monitor
            requestAnimationFrame(draw);
        });

        canvas.addEventListener('wheel', (e) => {
            e.preventDefault();
            scale += e.deltaY * -0.5;
            if (scale < 10) scale = 10;
            requestAnimationFrame(draw);
        }, { passive: false });

        // Fungsi utama untuk kalkulasi matematika 3D dan menggambar ke Canvas
        function draw() {
            ctx.clearRect(0, 0, width, height);
            ctx.strokeStyle = '#2c3e50';
            ctx.lineWidth = 1.0;

            const sinX = Math.sin(rx);
            const cosX = Math.cos(rx);
            const sinY = Math.sin(ry);
            const cosY = Math.cos(ry);

            const cx = width / 2;
            const cy = height / 2;

            // Memproyeksikan array Vertices 3D ke kooordinat 2D
            const projected = model.Vertices.map(v => {
                const x1 = v.X * cosY - v.Z * sinY;
                const z1 = v.X * sinY + v.Z * cosY;
                const y1 = v.Y;
                const y2 = y1 * cosX - z1 * sinX;
                return { x: x1 * scale + cx, y: -y2 * scale + cy };
            });

            // Menggambar garis wajah (faces) ke layar
            ctx.beginPath();
            model.Faces.forEach(face => {
                if (face.length < 3) return;
                ctx.moveTo(projected[face[0]].x, projected[face[0]].y);
                for (let i = 1; i < face.length; i++) {
                    ctx.lineTo(projected[face[i]].x, projected[face[i]].y);
                }
                ctx.closePath();
            });
            ctx.stroke();
        }

        // Render pertama kali saat halaman dimuat
        draw();
    </script>
</body>
</html>`

	// Menyuntikkan JSON ke dalam template HTML
	t := template.Must(template.New("index").Parse(html))
	t.Execute(w, struct{ ModelData template.JS }{template.JS(modelJSON)})
}
