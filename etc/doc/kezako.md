## peerpx c'est quoi donc ?  
  
Peerpx est un "service" de partage de photos (comme flickr ou 500px) décentralisé et (eventuellement) fédéré.  
  
### Noeud  
  
Un noeud est une instance autonome qui propose toutes les fonctionnalités nécessaires à un service complet de partage de photos.  
  
Un utilisateur (non admin) pourra se créer un compte sur n'importe quel noeud (à condition que l'admin soit OK). Son identifiant global sera alors @user@node.  
  
Un utilisateur sur un noeud X devra pouvoir suivre un utilisateur du noeud Z même si les deux noeuds ne sont pas fédérés.  
  
  
  
### Fédération  
  
Chaque noeud doit pouvoir se fédérer avec X autres noeuds. Ainsi depuis n'importe quel noeud d'une fédération il sera possible de consulter les photos publiées sur n'importe quel noeud fédéré.  
Il faudra proposer une fédération "principale" mais chacun sera libre d'y adhérer ou pas.  
  
  
## Technique  
  
### Frontend  
  
-> vuejs ou equivalent ?  
  
### Backend  
  
En Go bien entendu ;)  
  
Idéalement il faudrait s'arranger pour que ce soit dispo sur du multiplateforme (pour pouvoir proposer par exemple une image disque pour raspberry) donc éviter les package dépendant de lib C non disponnibles sous Win et ARM.
  
### DB  
Vu que chaque noeud va avoir besoin de ressources différentes en fonction du nombre d'utilisateurs/photo/fede, il faut abstraire au maximum l'interaction avec la DB pour pouvoir utiliser plusieurs "moteurs".
  
Par exemple pour un noeud qui aura peu d'utilisateurs et/ou peu de photos il serait dommage de déployer du MySQL ou PostgreSQL alors que SQLite serait beaucoup plus approprié.  

On pourrait utiliser un ORM comme [GORM](http://gorm.io/)    
  
Par contre le truc ennuyeux c'est que chaque noeud va devoir stocker les datas de ce qu'il héberge (contenus et utilisateurs) mais aussi le contenu des noeuds fédérés et de leurs utilisateurs pour avoir des réponses rapides en cas de recherches. Et ça risque d'être difficilement gérable....
Avant de partir vers ce genre de solution il faudra tester ce que ça donne en balançant les requêtes en parallèle sur les noeud fédérés, mais la aussi j'ai des doutes. Imaginez par exemple un noeud de la  fédération principale, si le truc prends, il va devoir faire des centaines de requêtes vers autant de noeuds, attendre lles résultats et les agréger... Bref il va falloir se creuser les méninges sur ce point.
  
### Activitypub  
  
Activitypub semble être le protocole le plus approprié pour gérer les différents événements (nouvelle photo, commentaire, note, ...) entre utilisateurs et entre noeuds.  
  
Pour avoir une idée un peu plus précise lisez l'intro! [https://www.w3.org/TR/2018/REC-activitypub-20180123/](https://www.w3.org/TR/2018/REC-activitypub-20180123/)  
  
 
### Stockage  
Chaque noeud servira de stockage partagé via IPFS aux noeuds appartenant à la même fédération (ou aux mêmes).  
Autrement dit quand on consultera une photo sur le noeud X elle sera chargée depuis une noeud appartenant à la même fédération. Il faudra donc prévoir un bidule qui répartit les requêtes et qui tient compte de la BP dispo sur chaque noeud (on peut aussi imaginer de servir le client depuis le noeud le plus proche.. mais bon ce n'est pas le sujet pour le moment)