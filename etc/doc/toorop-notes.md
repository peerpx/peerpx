## ActivityPub pour madame Michu  
  
Quand j'attaque un nouveau concept, j'aime bien, pour mieux l'appréhender, le regarder de très haut et ne garder que l'essentiel, donc voici ma vision actuelle d'activityPub:     

On va prendre comme exemple ce que l'on connait tous: Twitter.  
  
Soit:  
  
- les utilisateurs @user1 et @user2  
- user2 suit user1  
  
  
Pour chaque user:  
  
- Sa INBOX contient le flux de tweets de tous les users qu'il suit. Concrétement c'est sa page d'accueil twitter -> https://twitter.com/    
  
- Sa OUTBOX contient uniquement son flux ->  https://twitter.com/user  
   
  
Quand user1 poste un tweet:  
  
- Ce tweet (une référence à) va dans sa OUTBOX  
- Ce tweet (une référence à) va dans la INBOX de user2 (et dans toutes les INBOX des users qui le suivent)  
  
  
## Acteurs activityPub  
  
- user  
- groupe (des galeries partagées ayant un thème spécifique)  
- album (galerie propre a un user qui peut être suivie par d'autres users)  
  
- instance (au sens large autrement dit le @instance de @user@instance). 
  
  
## App mobile  
- pour moi c'est indispensable (ne serait ce que pour l'usage "instagram" (voir plus bas))
- flutter ?  
  
## Usages du service  
Idéalement ça doit couvrir les deux extremes:  
  
- Instagram:  tout et n'importe quoi  
- ce qu'était 500px à l'origine: reservé au photographes pour y poster leurs plus belles photos. Donc très qualitatif.  

Un truc ton con pour qu'une instance soit orientée dans un sens ou dans l'autre c'est de limiter le nombre d'images qu'un user peut poster par période. 
  
  
## Reproches fait au services actuels auxquels il faut apporter des solutions  
  
[AMA du CEO de 500px qui expose ces problemes](https://www.reddit.com/r/photography/comments/66cbpa/hey_reddit_im_andy_yang_ceo_of_500px_ready_to/)  
  
- loop(tu as beaucoup de followers -> beaucoup de "like/comment" par photo postée -> plus de visibilté -> plus de followers).  
Une idée en passant pour l'algo de mise en avant, en plus des likes/comments on pourrait mesurer le temps passé par les autres users à regarder la photo (et en particulier le temps moyen par user). C'est facilement contournable par un bot mais c'est assez coûteux si le bot doit passer X secondes sur chaque photo qu'il veut mettre en avant. (reste a voir si une implémentation fiable est possible) 
   
- les bots -> limiter les appels API par user (au moins pour les likes, comments, follow...)