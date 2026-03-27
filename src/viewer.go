package main

import (
	"encoding/json"
	"html/template"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Model struct {
	Vertices []Vector3
	Faces    [][]int
}

type VoxelizeResponse struct {
	Model   Model  `json:"model"`
	ObjData string `json:"objData"`
}

var currentModel Model

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

func handleVoxelize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, "Gagal memproses form", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Gagal mendapatkan file upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	depthStr := r.FormValue("depth")
	maxDepth, err := strconv.Atoi(depthStr)
	if err != nil || maxDepth <= 0 {
		maxDepth = 5
	}

	tempInput := "temp_input.obj"
	tempOutput := "temp_output.obj"

	outFile, err := os.Create(tempInput)
	if err != nil {
		http.Error(w, "Gagal menyimpan file temporary", http.StatusInternalServerError)
		return
	}
	io.Copy(outFile, file)
	outFile.Close()

	startTime := time.Now()

	vertices, faces, rootBoundary, err := ParseOBJ(tempInput)
	if err != nil {
		http.Error(w, "Gagal parsing OBJ", http.StatusInternalServerError)
		return
	}

	rootNode := NewOctreeNode(rootBoundary)
	stats := NewOctreeStats(maxDepth)

	var wg sync.WaitGroup
	wg.Add(1)
	BuildOctreeConcurrent(rootNode, vertices, faces, 0, maxDepth, stats, &wg)
	wg.Wait()

	var solidVoxels []BoundingBox
	CollectVoxels(rootNode, &solidVoxels)

	// --- AWAL LOGIKA OPTIMASI PERMUKAAN (HIDDEN FACE CULLING) ---
	gridDim := 1 << maxDepth
	rootSize := rootBoundary.Max.X - rootBoundary.Min.X
	voxelSize := rootSize / float64(gridDim)

	om := NewOptimizationMap(solidVoxels, rootBoundary.Min, voxelSize)

	optimizedModel := Model{Vertices: []Vector3{}, Faces: [][]int{}}
	vertexCount := 0

	for key := range om.VoxelGrid {
		voxelMinX := float64(key.X)*voxelSize + rootBoundary.Min.X
		voxelMinY := float64(key.Y)*voxelSize + rootBoundary.Min.Y
		voxelMinZ := float64(key.Z)*voxelSize + rootBoundary.Min.Z

		vMin := Vector3{X: voxelMinX, Y: voxelMinY, Z: voxelMinZ}
		vMax := Vector3{X: voxelMinX + voxelSize, Y: voxelMinY + voxelSize, Z: voxelMinZ + voxelSize}

		cubeVertices := []Vector3{
			{X: vMin.X, Y: vMin.Y, Z: vMin.Z}, {X: vMax.X, Y: vMin.Y, Z: vMin.Z}, {X: vMax.X, Y: vMin.Y, Z: vMax.Z}, {X: vMin.X, Y: vMin.Y, Z: vMax.Z},
			{X: vMin.X, Y: vMax.Y, Z: vMin.Z}, {X: vMax.X, Y: vMax.Y, Z: vMin.Z}, {X: vMax.X, Y: vMax.Y, Z: vMax.Z}, {X: vMin.X, Y: vMax.Y, Z: vMax.Z},
		}

		offset := vertexCount
		optimizedModel.Vertices = append(optimizedModel.Vertices, cubeVertices...)

		// Cek sisi yang bersentuhan, buang sisi dalam
		if !om.IsInternalFace(key.X, key.Y, key.Z, "Bawah") {
			optimizedModel.Faces = append(optimizedModel.Faces, []int{offset, offset + 1, offset + 2, offset + 3})
		}
		if !om.IsInternalFace(key.X, key.Y, key.Z, "Atas") {
			optimizedModel.Faces = append(optimizedModel.Faces, []int{offset + 4, offset + 5, offset + 6, offset + 7})
		}
		if !om.IsInternalFace(key.X, key.Y, key.Z, "Belakang") {
			optimizedModel.Faces = append(optimizedModel.Faces, []int{offset, offset + 1, offset + 5, offset + 4})
		}
		if !om.IsInternalFace(key.X, key.Y, key.Z, "Depan") {
			optimizedModel.Faces = append(optimizedModel.Faces, []int{offset + 2, offset + 3, offset + 7, offset + 6})
		}
		if !om.IsInternalFace(key.X, key.Y, key.Z, "Kiri") {
			optimizedModel.Faces = append(optimizedModel.Faces, []int{offset + 3, offset, offset + 4, offset + 7})
		}
		if !om.IsInternalFace(key.X, key.Y, key.Z, "Kanan") {
			optimizedModel.Faces = append(optimizedModel.Faces, []int{offset + 1, offset + 2, offset + 6, offset + 5})
		}
		vertexCount += 8
	}

	currentModel = normalizeModel(optimizedModel)
	// --- AKHIR LOGIKA OPTIMASI ---

	// Export voxel utuh ke file .obj untuk fitur Download
	numVertices, numFaces, err := ExportToObj(solidVoxels, tempOutput)
	if err != nil {
		http.Error(w, "Gagal export voxel", http.StatusInternalServerError)
		return
	}

	executionTime := time.Since(startTime)
	PrintReport(stats, len(solidVoxels), numVertices, numFaces, maxDepth, executionTime, fileHeader.Filename+" (Voxelized)")

	objBytes, err := os.ReadFile(tempOutput)
	var rawObjData string
	if err == nil {
		rawObjData = string(objBytes)
	}

	os.Remove(tempInput)
	os.Remove(tempOutput)

	response := VoxelizeResponse{
		Model:   currentModel,
		ObjData: rawObjData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	modelJSON, err := json.Marshal(currentModel)
	if err != nil {
		modelJSON = []byte("{}")
	}

	html := `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <title>Go 3D Object Viewer - Solid Voxel</title>
    <style>
        body { margin: 0; overflow: hidden; background-color: #f4f4f9; font-family: sans-serif; }
        canvas { display: block; cursor: grab; }
        canvas:active { cursor: grabbing; }
        .controls { position: absolute; top: 10px; left: 10px; background: rgba(255,255,255,0.9); padding: 15px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); width: 250px; }
        input[type="file"], input[type="number"], button { width: 100%; margin-top: 5px; box-sizing: border-box; }
        button { padding: 8px; background: #2c3e50; color: white; border: none; border-radius: 4px; cursor: pointer; margin-top: 15px; font-weight: bold; transition: background 0.2s; }
        button:hover { background: #34495e; }
        .btn-download { background: #27ae60; margin-top: 10px; display: none; }
        .btn-download:hover { background: #219653; }
    </style>
</head>
<body>
    <div class="controls">
        <h3 style="margin-top:0;">Go 3D Viewer</h3>
        <p style="margin-bottom:10px; font-size: 13px;">Tahan klik dan geser untuk <b>Rotasi</b>.<br>Gunakan scroll untuk <b>Zoom</b>.</p>
        <hr style="border: 0; border-top: 1px solid #ccc; margin: 10px 0;">
        
        <form id="uploadForm">
            <label style="font-size: 14px; font-weight: bold;">Upload File .obj:</label>
            <input type="file" id="objFile" accept=".obj" required>
            
            <label style="font-size: 14px; font-weight: bold; display: block; margin-top: 10px;">Max Depth Octree:</label>
            <input type="number" id="depth" value="5" min="1" max="8">
            
            <button type="submit">Proses Voxelization</button>
        </form>
        
        <div id="loading" style="display: none; margin-top: 15px; color: #d35400; font-weight: bold; font-size: 14px; text-align: center;">
            Memproses Voxelization...<br><small>Mohon tunggu sebentar</small>
        </div>

        <button type="button" id="downloadBtn" class="btn-download">💾 Simpan Hasil .obj</button>
    </div>
    
    <canvas id="viewer"></canvas>

    <script>
        let model = {{.ModelData}};
        let currentObjData = ""; 
        let currentFileName = "voxelized.obj"; 

        const canvas = document.getElementById('viewer');
        const ctx = canvas.getContext('2d');

        let width = window.innerWidth;
        let height = window.innerHeight;
        canvas.width = width;
        canvas.height = height;

        window.addEventListener('resize', () => {
            width = window.innerWidth;
            height = window.innerHeight;
            canvas.width = width;
            canvas.height = height;
            draw();
        });

        let rx = 0.5; let ry = -0.5;
        let scale = Math.min(width, height) / 3;
        let isDragging = false;
        let lastX = 0; let lastY = 0;

        canvas.addEventListener('mousedown', (e) => { isDragging = true; lastX = e.clientX; lastY = e.clientY; });
        window.addEventListener('mouseup', () => { isDragging = false; });

        canvas.addEventListener('mousemove', (e) => {
            if (!isDragging) return;
            ry += (e.clientX - lastX) * 0.01;
            rx += (e.clientY - lastY) * 0.01;
            lastX = e.clientX; lastY = e.clientY;
            requestAnimationFrame(draw);
        });

        canvas.addEventListener('wheel', (e) => {
            e.preventDefault();
            scale += e.deltaY * -0.5;
            if (scale < 10) scale = 10;
            requestAnimationFrame(draw);
        }, { passive: false });

        function draw() {
            ctx.clearRect(0, 0, width, height);
            
            if (!model || !model.Vertices || model.Vertices.length === 0) {
                ctx.fillStyle = '#7f8c8d';
                ctx.font = 'bold 20px sans-serif';
                ctx.textAlign = 'center';
                ctx.fillText('Silakan unggah file .obj melalui form di kiri.', width / 2, height / 2);
                return;
            }

            // Garis pinggir yang tipis dan agak transparan
            ctx.strokeStyle = 'rgba(44, 62, 80, 0.7)';
            ctx.lineWidth = 0.3;

            const sinX = Math.sin(rx); const cosX = Math.cos(rx);
            const sinY = Math.sin(ry); const cosY = Math.cos(ry);
            const cx = width / 2; const cy = height / 2;

            // 1. Proyeksi 3D ke 2D dan simpan kedalaman Z (Z-Depth)
            const projected = model.Vertices.map(v => {
                const x1 = v.X * cosY - v.Z * sinY;
                const z1 = v.X * sinY + v.Z * cosY;
                const y1 = v.Y;
                const y2 = y1 * cosX - z1 * sinX;
                const z2 = z1 * cosX + y1 * sinX; // Kedalaman Z dihitung di sini
                return { x: x1 * scale + cx, y: -y2 * scale + cy, z: z2 };
            });

            // 2. Kumpulkan semua sisi (faces) beserta rata-rata kedalaman Z-nya
            const facesToDraw = [];
            model.Faces.forEach(face => {
                if (face.length < 3) return;
                let zSum = 0;
                for (let i = 0; i < face.length; i++) {
                    zSum += projected[face[i]].z;
                }
                facesToDraw.push({
                    indices: face,
                    zDepth: zSum / face.length
                });
            });

            // 3. Painter's Algorithm: Urutkan dari sisi paling belakang ke paling depan
            facesToDraw.sort((a, b) => a.zDepth - b.zDepth);

            // 4. Gambar dan warnai secara berurutan
            facesToDraw.forEach(faceObj => {
                const face = faceObj.indices;
                ctx.beginPath();
                ctx.moveTo(projected[face[0]].x, projected[face[0]].y);
                for (let i = 1; i < face.length; i++) {
                    ctx.lineTo(projected[face[i]].x, projected[face[i]].y);
                }
                ctx.closePath();
                
                // Isi persegi dengan warna solid untuk MENUTUPI garis yang ada di belakangnya
                ctx.fillStyle = '#ecf0f1'; // Warna abu-abu cerah ala voxel
                ctx.fill();
                
                // Gambar garis tepi persegi
                ctx.stroke();
            });
        }

        document.getElementById('uploadForm').addEventListener('submit', function(e) {
            e.preventDefault();
            
            const fileInput = document.getElementById('objFile');
            const depthInput = document.getElementById('depth');
            const loadingDiv = document.getElementById('loading');
            const downloadBtn = document.getElementById('downloadBtn');
            
            if (fileInput.files.length === 0) return;
            
            let originalName = fileInput.files[0].name;
            currentFileName = originalName.replace('.obj', '') + '_voxelized.obj';
            
            const formData = new FormData();
            formData.append('file', fileInput.files[0]);
            formData.append('depth', depthInput.value);
            
            loadingDiv.style.display = 'block';
            downloadBtn.style.display = 'none'; 
            
            fetch('/voxelize', {
                method: 'POST',
                body: formData
            })
            .then(response => {
                if (!response.ok) throw new Error("Gagal memproses file");
                return response.json();
            })
            .then(data => {
                model = data.model; 
                currentObjData = data.objData;
                
                rx = 0.5; ry = -0.5;
                scale = Math.min(width, height) / 3;
                draw(); 
                
                loadingDiv.style.display = 'none';
                downloadBtn.style.display = 'block'; 
            })
            .catch(err => {
                alert(err.message);
                loadingDiv.style.display = 'none';
            });
        });

        document.getElementById('downloadBtn').addEventListener('click', function() {
            if (!currentObjData) return;
            
            const blob = new Blob([currentObjData], { type: 'text/plain' });
            const url = URL.createObjectURL(blob);
            
            const a = document.createElement('a');
            a.href = url;
            a.download = currentFileName;
            document.body.appendChild(a);
            a.click();
            
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
        });

        draw();
    </script>
</body>
</html>`

	t := template.Must(template.New("index").Parse(html))
	t.Execute(w, struct{ ModelData template.JS }{template.JS(modelJSON)})
}
