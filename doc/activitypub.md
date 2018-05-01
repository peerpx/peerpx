# ActivityPub

ActivityPub est un "protocole" d'échange de données pour la construction d'un/de réseau(x) social/aux décentralisé. Il est basé sur le format de données ActivityStream (JSON-LD) 2.0.

## Explications, détails, (et considérations)

L'usage d'HTTPS en vivement encouragé pour tous les échanges.

**2 types de communication :**
  * Client - Serveur
  * Serveur - Serveur (fédération)

Des *acteur*s interagissent en échangeant des *activités* sur des *objet*s (Create, Update, ...) (format ActivityStream, un *acteur* est un *objet*, une *activité* concerne un *objet*).

Plusieurs *utilisateur*s peuvent "gérer" un *acteur* et inversement un *utilisateur* peut gérer plusieurs *acteur*s

Chaque *acteur* possède une INBOX et une OUTBOX, qui sont **GET**able et **POST**able comme suit :
  * INBOX
    * POST - cas de la fédération, un serveur envoie à un *acteur* les *activité*s qui lui sont destinées
    * GET - le client de l'*acteur* récupère les *activité*s qui lui sont destinées
  * OUTBOX
    * POST - le client de l'*acteur* envoie un *activité* destinée à une/des entités externe(s)
    * GET - la/les entité(s) externes (serveur) récupèrent les *activité*s qui leurs sont destinées

Chaque *acteur* devrait aussi avoir une collection de "followers" et de "following", et peut avoir une collection de "liked" (*objet*s aimés)