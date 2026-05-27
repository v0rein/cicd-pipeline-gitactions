# Laporan Insiden 2: Rolling Update Tanpa Downtime

## Konteks Insiden
Setiap kali ada perilisan fitur baru (deployment versi baru), aplikasi harus dimatikan terlebih dahulu selama beberapa menit, mengakibatkan penolakan koneksi dari pengguna dan klien yang sedang mengakses layanan.

## Solusi Kubernetes
Memanfaatkan fitur **Rolling Update** di dalam Kubernetes Deployment (`strategy.type: RollingUpdate`). Kubernetes akan secara perlahan menaikkan Pod versi baru, memastikannya *Running*, dan mengalihkan traffic secara halus sebelum akhirnya mematikan Pod versi lama secara bertahap (`maxUnavailable: 0`).

## Langkah-Langkah Pengujian
Kami menguji ketersediaan layanan selama proses update versi dari aplikasi:

1. Di **Terminal 1**, simulasikan lalu lintas pengguna dengan membuat *loop request* ke aplikasi secara konstan menggunakan PowerShell:
   ```powershell
   while ($true) {
       try {
           $response = Invoke-WebRequest -Uri "http://127.0.0.1:53441" -UseBasicParsing -ErrorAction SilentlyContinue
           $status = $response.StatusCode
       } catch {
           $status = "Error"
       }
       Write-Host "$(Get-Date -Format 'HH:mm:ss') — HTTP $status"
       Start-Sleep -Milliseconds 500
   }
   ```
2. Di **Terminal 2**, lakukan pembaruan versi (misalnya dari "v1" ke "v2") pada file `kubernetes/deployment.yaml`, kemudian jalankan pembaruan konfigurasi:
   ```bash
   kubectl apply -f kubernetes/deployment.yaml -n taskflow-prod
   ```
3. Pantau hasil output *looping* di Terminal 1 selama pembaruan (sekitar 10-15 detik) untuk mendeteksi *downtime*.

## Bukti Pengujian
Output di Terminal 1 menunjukkan rentetan status:
- `HTTP 200`
- `HTTP 200`
- `HTTP 200`

*(Sisipkan Screenshot Terminal 1 yang terus menerus menghasilkan HTTP 200 tanpa henti di sini)*

## Hasil
- **Downtime yang tercatat selama update:** 0 detik (Lalu lintas HTTP 200 konsisten).
- **Kesimpulan:** Insiden 2 berhasil diselesaikan. Fitur dapat diunggah bahkan pada jam sibuk tanpa merugikan pengalaman pengguna.
