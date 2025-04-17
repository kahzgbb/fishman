# FishMan

The program detects files deleted over a period of time and displays how long ago they were deleted and when they were executed.
<br></br>

## What is used?
```
C:\Windows\Prefetch
```
- The prefetch method is used to check the files that have been executed on the computer.
<br></br>
```
SYSTEM\CurrentControlSet\Control\Session Manager\AppCompatCache
```

- The AppCompactCache regedit path is also used to check the files that have passed through the computer.
<br></br>

```
C:\Windows\AppCompat\Programs\Amcache.hve
```

- Amcache is used to compare and obtain the file execution date.
<br></br>
## ⚙ Requirements

- Windows 10+
- Disable Windows Defender (or the antivirus if it blocks the program)
- Run the program as administrator
<br></br>
## ⚠ Disclaimer

Some detections may be false (there may be a small chance), but most are true and if anything happens, just check the found file.
<br></br>
## 🎫 License

This project is subject to [GNU General Public License v3.0](LICENSE).
