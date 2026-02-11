Vamos a hacer un refactor del cli:

- [x] agregar el --help para mostrar todos los comandos disponibles
- [x] Mover el agregar skills a 'skli add [git-repo-path]' en el que se le puede pasar opcionalmente el git repo path si no se muestra el tui con el input
- [x] La eliminación de skills se hará con 'skli rm [skill-name]'  si no se da el name se muestra el listado de skills en la carpeta de trabajo local y checkbox para ir marcando que skills quitar
- [x] skli sync 'descarga los skills que han cambiado
- [x] skli upload [git-dest-repo-path] [local-skill-path] Si no se indica se debe gestionarse desde el tui en dos pasos, el primero indicar el repo y el segundo mostrará un listado de todos los skills locales no sincronizados. 
- [x] skli config. Se queda igual mostrando el listado de origins y el de carpetas para guardar los skills

Ve marcando con checks cuando vayas agregando cada funcionalidad
