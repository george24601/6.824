1. ps -a  | sort -k 4
2. grep -vi '[aeiou]' /usr/share/dict/words | wc
3. yes "1 0 " | fmt 9 | head -5
4. 
```
ls -laS /etc/*.conf
```
5. the second command will give the permission issue. Race condition for the file
6. temp4 is the case of race conditon -> yynn instead of ynyn as in temp3
7. To be compatible between devices, so that there is no separate mode between seq and random IO. Use seek to locate the r/w pointer to the specific offset
8. same dir; parent dir
9. grep the current dir
10. access time changed 
11. link count for increased by 1, because the new folder has a .. link
12. easy to get file locations without copying. but symbolic links points to the name instead of inode. Therefore, if the file is renamed, we can not access the content in that snaymore  symblic name
17. file descriptor
