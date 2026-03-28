# Voxelization Objek 3D menggunakan Octree 🧊

Program ini adalah implementasi algoritma **Divide and Conquer** menggunakan struktur data **Octree** untuk mengonversi model 3D (Polygon Mesh berformat `.obj`) menjadi model **Voxel** (kumpulan kubus). Program ini dibangun menggunakan bahasa **Go (Golang)** dan dilengkapi dengan visualisasi 3D interaktif berbasis Web (Client-Side Rendering) tanpa menggunakan *library* grafis eksternal.

Tugas ini dikerjakan untuk memenuhi Tugas Kecil 2 Mata Kuliah IF2211 Strategi Algoritma, Program Studi Teknik Informatika, Institut Teknologi Bandung.

## 👨‍💻 Identitas Pembuat
* **13524036** - Edward David Rumahorbo
* **13524044** - Narendra Dharma Wistara M.

## ✨ Fitur Program
1. **Octree Voxelization:** Membagi ruang 3D menjadi 8 oktan secara rekursif (Divide & Conquer) hingga batas kedalaman (`maxDepth`) yang ditentukan.
2. **SAT (Separating Axis Theorem):** Deteksi persinggungan presisi tingkat tinggi antara poligon objek dengan kubus *node* untuk mencegah voxel berlebih (*false positive*).
3. **Concurrency (Bonus):** Pemrosesan rekursif Octree dioptimasi menggunakan *Goroutines* secara paralel agar eksekusi jauh lebih cepat.
4. **Interactive Web Viewer (Bonus):** Visualisasi model 3D langsung di dalam *browser* menggunakan HTML5 `<canvas>`, lengkap dengan fitur rotasi (klik & geser) dan *zoom* (*scroll*).
5. **Hidden Face Culling & Painter's Algorithm:** Optimasi visual agar kubus terlihat solid (padat) dan garis-garis yang berada di dalam objek/tertutup tidak perlu dirender, sehingga performa *browser* tetap ringan.
6. **Export & Download:** Fitur untuk menyimpan hasil voxelization menjadi file `.obj` baru langsung melalui antarmuka Web.

## 🛠️ Persyaratan Sistem (Dependensi)
Karena program ini dirancang agar portabel dan murni menggunakan *Standard Library* Go, Anda hanya memerlukan:
* **Go Compiler:** Versi 1.18 atau yang lebih baru (untuk *build* atau *run*).
* **Web Browser Modern:** Chrome, Firefox, Edge, atau Safari (untuk membuka antarmuka visualisasi).
* *(Tidak ada library eksternal atau CGO yang perlu diinstal).*

## 🚀 Cara Instalasi (Clone Repository)
Buka terminal/Command Prompt Anda dan jalankan perintah berikut untuk mengunduh repositori ini ke komputer lokal:

```bash
git clone https://github.com/bangedaw/Tucil2_13524036_13524044.git
cd Tucil2_13524036_13524044
```

## 💻 Cara Menjalankan Program

### ▶️ Untuk Pengguna Windows / macOS
Bagi pengguna Windows, Anda dapat menjalankan *source code* secara langsung tanpa perlu melakukan kompilasi manual. 
1. Masuk ke dalam direktori `src`.
2. Jalankan perintah `go run .`

```cmd
cd src
go run .
```

### ▶️ Untuk Pengguna Linux
Bagi pengguna Linux atau macOS, program dijalankan melalui *executable file* yang sudah disediakan di dalam folder `bin`.
1. Masuk ke dalam direktori `bin`.
2. Eksekusi file binari `voxelizer`.

```bash
cd bin
./voxelizer
```

*(Catatan: Jika file executable di folder `bin` tidak dapat dijalankan karena perbedaan arsitektur OS, silakan lakukan build ulang dengan masuk ke folder `src` dan jalankan perintah: `go build -o ../bin/voxelizer .` lalu ulangi langkah di atas).*

## 🎮 Cara Menggunakan Program & Visualisasi
1. Setelah program dijalankan melalui terminal, server lokal akan aktif. Jangan tutup terminal tersebut.
2. Buka *Web Browser* Anda dan akses alamat berikut: **`http://localhost:8080`**
3. Pada panel antarmuka di sebelah kiri layar:
   * Klik **"Pilih File"** atau **"Browse"** untuk mengunggah file `.obj` mentah. (Contoh file uji tersedia di dalam folder `test/input/` seperti `cow.obj`, `pumpkin.obj`, atau `teapot.obj`).
   * Atur **"Max Depth Octree"** (Batas kedalaman rekursi). Nilai *default* adalah 5. *(Catatan: Nilai 6 ke atas akan menghasilkan resolusi voxel yang sangat detail namun membutuhkan waktu komputasi yang lebih lama).*
   * Klik tombol **"Proses Voxelization"**.
4. Tunggu beberapa saat. Laporan komputasi (jumlah voxel, waktu eksekusi, dll) akan otomatis tercetak di jendela **Terminal** Anda.
5. Setelah selesai, model Voxel 3D akan muncul di layar *browser*.
   * **Rotasi:** Tahan Klik Kiri pada *mouse* dan geser ke segala arah.
   * **Zoom:** Gunakan *Scroll Wheel* pada *mouse*.
6. Jika hasil sudah sesuai, Anda bisa mengklik tombol hijau **"💾 Simpan Hasil .obj"** untuk mengunduh file `.obj` voxelization tersebut ke komputer Anda.
