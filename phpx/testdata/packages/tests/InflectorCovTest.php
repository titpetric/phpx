<?php
use Doctrine\Inflector\{InflectorFactory, Inflector, Language, RulesetInflector, CachedWordInflector, NoopWordInflector, WordInflector};
use Doctrine\Inflector\Rules\{Pattern, Patterns, Ruleset, Substitution, Substitutions, Transformation, Transformations, Word};
use Doctrine\Inflector\Rules;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

final class InflectorCovTest extends TestCase
{
    private const DATA = <<<'JSON'
{"english":[["user","users","user"],["person","people","person"],["box","boxes","box"],["city","cities","city"],["mouse","mice","mouse"],["child","children","child"],["analysis","analyses","analysi"],["quiz","quizzes","quiz"],["ox","oxen","ox"],["bus","buses","bu"],["day","days","day"],["lady","ladies","lady"],["knife","knives","knife"],["wolf","wolves","wolf"],["man","men","man"],["woman","women","woman"],["tooth","teeth","tooth"],["foot","feet","foot"],["cactus","cacti","cactu"],["criterion","criteria","criterion"],["matrix","matrices","matrix"],["vertex","vertices","vertex"],["cat","cats","cat"],["dog","dogs","dog"],["house","houses","house"],["cow","cows","cow"],["potato","potatoes","potato"],["hero","heroes","hero"],["status","statuses","status"],["news","news","news"],["animal","animals","animal"],["mal","mals","mal"],["cidadão","cidadãos","cidadão"],["pão","pãos","pão"],["país","país","paí"],["mês","mês","mê"],["lápis","lápis","lápi"],["cavalo","cavalos","cavalo"],["flor","flors","flor"],["papel","papels","papel"],["canal","canals","canal"],["luz","luzs","luz"],["homem","homems","homem"],["alemão","alemãos","alemão"],["livro","livros","livro"],["ônibus","ônibuses","ônibu"],["cantor","cantors","cantor"],["amigo","amigos","amigo"],["amiga","amigas","amiga"],["giorno","giornos","giorno"],["uomo","uomos","uomo"],["città","cittàs","città"],["libro","libros","libro"],["lago","lagos","lago"],["amico","amicos","amico"],["dio","dios","dio"],["gatto","gattos","gatto"],["bambino","bambinos","bambino"],["casa","casas","casa"],["hundo","hundos","hundo"],["katt","katts","katt"],["bok","boks","bok"],["mann","manns","mann"],["dag","dags","dag"],["bil","bils","bil"],["gutt","gutts","gutt"],["barn","barns","barn"],["arm","arms","arm"],["boek","boeks","boek"],["estudiante","estudiantes","estudiante"],["lápiz","lápizs","lápiz"],["corazón","corazóns","corazón"],["ciudad","ciudads","ciudad"],["mes","mes","me"],["rey","reys","rey"],["mujer","mujers","mujer"],["canción","cancións","canción"],["ej","ejs","ej"],["el","els","el"],["elma","elmas","elma"],["kitap","kitaps","kitap"],["göz","gözs","göz"],["ev","evs","ev"],["çocuk","çocuks","çocuk"],["adam","adams","adam"],["kadın","kadıns","kadın"],["araba","arabas","araba"],["okul","okuls","okul"],["monato","monatos","monato"],["knabo","knabos","knabo"],["hundoj","hundojs","hundoj"],["katoj","katojs","katoj"],["chat","chats","chat"],["cheval","chevals","cheval"],["journal","journals","journal"],["animaux","animauxes","animaux"],["chou","chous","chou"],["bijou","bijous","bijou"],["travail","travails","travail"],["oeil","oeils","oeil"],["nez","nezs","nez"],["pays","pays","pay"],["roi","rois","roi"],["prix","prixes","prix"],["feu","feus","feu"],["lieu","lieus","lieu"],["os","os","o"],["œuf","œufs","œuf"],["jeu","jeus","jeu"]],"esperanto":[["user","user","user"],["person","person","person"],["box","box","box"],["city","city","city"],["mouse","mouse","mouse"],["child","child","child"],["analysis","analysis","analysis"],["quiz","quiz","quiz"],["ox","ox","ox"],["bus","bus","bus"],["day","day","day"],["lady","lady","lady"],["knife","knife","knife"],["wolf","wolf","wolf"],["man","man","man"],["woman","woman","woman"],["tooth","tooth","tooth"],["foot","foot","foot"],["cactus","cactus","cactus"],["criterion","criterion","criterion"],["matrix","matrix","matrix"],["vertex","vertex","vertex"],["cat","cat","cat"],["dog","dog","dog"],["house","house","house"],["cow","cow","cow"],["potato","potatoj","potato"],["hero","heroj","hero"],["status","status","status"],["news","news","news"],["animal","animal","animal"],["mal","mal","mal"],["cidadão","cidadãoj","cidadão"],["pão","pãoj","pão"],["país","país","país"],["mês","mês","mês"],["lápis","lápis","lápis"],["cavalo","cavaloj","cavalo"],["flor","flor","flor"],["papel","papel","papel"],["canal","canal","canal"],["luz","luz","luz"],["homem","homem","homem"],["alemão","alemãoj","alemão"],["livro","livroj","livro"],["ônibus","ônibus","ônibus"],["cantor","cantor","cantor"],["amigo","amigoj","amigo"],["amiga","amiga","amiga"],["giorno","giornoj","giorno"],["uomo","uomoj","uomo"],["città","città","città"],["libro","libroj","libro"],["lago","lagoj","lago"],["amico","amicoj","amico"],["dio","dioj","dio"],["gatto","gattoj","gatto"],["bambino","bambinoj","bambino"],["casa","casa","casa"],["hundo","hundoj","hundo"],["katt","katt","katt"],["bok","bok","bok"],["mann","mann","mann"],["dag","dag","dag"],["bil","bil","bil"],["gutt","gutt","gutt"],["barn","barn","barn"],["arm","arm","arm"],["boek","boek","boek"],["estudiante","estudiante","estudiante"],["lápiz","lápiz","lápiz"],["corazón","corazón","corazón"],["ciudad","ciudad","ciudad"],["mes","mes","mes"],["rey","rey","rey"],["mujer","mujer","mujer"],["canción","canción","canción"],["ej","ej","ej"],["el","el","el"],["elma","elma","elma"],["kitap","kitap","kitap"],["göz","göz","göz"],["ev","ev","ev"],["çocuk","çocuk","çocuk"],["adam","adam","adam"],["kadın","kadın","kadın"],["araba","araba","araba"],["okul","okul","okul"],["monato","monatoj","monato"],["knabo","knaboj","knabo"],["hundoj","hundoj","hundo"],["katoj","katoj","kato"],["chat","chat","chat"],["cheval","cheval","cheval"],["journal","journal","journal"],["animaux","animaux","animaux"],["chou","chou","chou"],["bijou","bijou","bijou"],["travail","travail","travail"],["oeil","oeil","oeil"],["nez","nez","nez"],["pays","pays","pays"],["roi","roi","roi"],["prix","prix","prix"],["feu","feu","feu"],["lieu","lieu","lieu"],["os","os","os"],["œuf","œuf","œuf"],["jeu","jeu","jeu"]],"french":[["user","users","user"],["person","persons","person"],["box","box","box"],["city","citys","city"],["mouse","mouses","mouse"],["child","childs","child"],["analysis","analysis","analysi"],["quiz","quiz","quiz"],["ox","ox","ox"],["bus","bus","bu"],["day","days","day"],["lady","ladys","lady"],["knife","knifes","knife"],["wolf","wolfs","wolf"],["man","mans","man"],["woman","womans","woman"],["tooth","tooths","tooth"],["foot","foots","foot"],["cactus","cactus","cactu"],["criterion","criterions","criterion"],["matrix","matrix","matrix"],["vertex","vertex","vertex"],["cat","cats","cat"],["dog","dogs","dog"],["house","houses","house"],["cow","cows","cow"],["potato","potatos","potato"],["hero","heros","hero"],["status","status","statu"],["news","news","new"],["animal","animaux","animal"],["mal","maux","mal"],["cidadão","cidadãos","cidadão"],["pão","pãos","pão"],["país","país","paí"],["mês","mês","mê"],["lápis","lápis","lápi"],["cavalo","cavalos","cavalo"],["flor","flors","flor"],["papel","papels","papel"],["canal","canaux","canal"],["luz","luz","luz"],["homem","homems","homem"],["alemão","alemãos","alemão"],["livro","livros","livro"],["ônibus","ônibus","ônibu"],["cantor","cantors","cantor"],["amigo","amigos","amigo"],["amiga","amigas","amiga"],["giorno","giornos","giorno"],["uomo","uomos","uomo"],["città","cittàs","città"],["libro","libros","libro"],["lago","lagos","lago"],["amico","amicos","amico"],["dio","dios","dio"],["gatto","gattos","gatto"],["bambino","bambinos","bambino"],["casa","casas","casa"],["hundo","hundos","hundo"],["katt","katts","katt"],["bok","boks","bok"],["mann","manns","mann"],["dag","dags","dag"],["bil","bils","bil"],["gutt","gutts","gutt"],["barn","barns","barn"],["arm","arms","arm"],["boek","boeks","boek"],["estudiante","estudiantes","estudiante"],["lápiz","lápiz","lápiz"],["corazón","corazóns","corazón"],["ciudad","ciudads","ciudad"],["mes","mes","me"],["rey","reys","rey"],["mujer","mujers","mujer"],["canción","cancións","canción"],["ej","ejs","ej"],["el","els","el"],["elma","elmas","elma"],["kitap","kitaps","kitap"],["göz","göz","göz"],["ev","evs","ev"],["çocuk","çocuks","çocuk"],["adam","adams","adam"],["kadın","kadıns","kadın"],["araba","arabas","araba"],["okul","okuls","okul"],["monato","monatos","monato"],["knabo","knabos","knabo"],["hundoj","hundojs","hundoj"],["katoj","katojs","katoj"],["chat","chats","chat"],["cheval","chevaux","cheval"],["journal","journaux","journal"],["animaux","animaux","animau"],["chou","choux","chou"],["bijou","bijoux","bijou"],["travail","travaux","travail"],["oeil","oeils","oeil"],["nez","nez","nez"],["pays","pays","pay"],["roi","rois","roi"],["prix","prix","prix"],["feu","feux","feu"],["lieu","lieux","lieu"],["os","os","o"],["œuf","œufs","œuf"],["jeu","jeux","jeu"]],"italian":[["user","user","user"],["person","person","person"],["box","box","box"],["city","city","city"],["mouse","mousi","mousa"],["child","child","child"],["analysis","analysis","analysis"],["quiz","quiz","quiz"],["ox","ox","ox"],["bus","bus","bus"],["day","day","day"],["lady","lady","lady"],["knife","knifi","knifa"],["wolf","wolf","wolf"],["man","man","man"],["woman","woman","woman"],["tooth","tooth","tooth"],["foot","foot","foot"],["cactus","cactus","cactus"],["criterion","criterion","criterion"],["matrix","matrix","matrix"],["vertex","vertex","vertex"],["cat","cat","cat"],["dog","dog","dog"],["house","housi","housa"],["cow","cow","cow"],["potato","potati","potato"],["hero","heri","hero"],["status","status","status"],["news","news","news"],["animal","animal","animal"],["mal","mal","mal"],["cidadão","cidadãi","cidadão"],["pão","pãi","pão"],["país","país","país"],["mês","mês","mês"],["lápis","lápis","lápis"],["cavalo","cavali","cavalo"],["flor","flor","flor"],["papel","papel","papel"],["canal","canal","canal"],["luz","luz","luz"],["homem","homem","homem"],["alemão","alemãi","alemão"],["livro","livri","livro"],["ônibus","ônibus","ônibus"],["cantor","cantor","cantor"],["amigo","amighi","amigo"],["amiga","amige","amiga"],["giorno","giorni","giorno"],["uomo","uomini","uomo"],["città","città","città"],["libro","libri","libro"],["lago","laghi","lago"],["amico","amici","amico"],["dio","dei","dio"],["gatto","gatti","gatto"],["bambino","bambini","bambino"],["casa","case","casa"],["hundo","hundi","hundo"],["katt","katt","katt"],["bok","bok","bok"],["mann","mann","mann"],["dag","dag","dag"],["bil","bil","bil"],["gutt","gutt","gutt"],["barn","barn","barn"],["arm","arm","arm"],["boek","boek","boek"],["estudiante","estudianti","estudianta"],["lápiz","lápiz","lápiz"],["corazón","corazón","corazón"],["ciudad","ciudad","ciudad"],["mes","mes","mes"],["rey","rey","rey"],["mujer","mujer","mujer"],["canción","canción","canción"],["ej","ej","ej"],["el","el","el"],["elma","elme","elma"],["kitap","kitap","kitap"],["göz","göz","göz"],["ev","ev","ev"],["çocuk","çocuk","çocuk"],["adam","adam","adam"],["kadın","kadın","kadın"],["araba","arabe","araba"],["okul","okul","okul"],["monato","monati","monato"],["knabo","knabi","knabo"],["hundoj","hundoj","hundoj"],["katoj","katoj","katoj"],["chat","chat","chat"],["cheval","cheval","cheval"],["journal","journal","journal"],["animaux","animaux","animaux"],["chou","chou","chou"],["bijou","bijou","bijou"],["travail","travail","travail"],["oeil","oeil","oeil"],["nez","nez","nez"],["pays","pays","pays"],["roi","roi","roe"],["prix","prix","prix"],["feu","feu","feu"],["lieu","lieu","lieu"],["os","os","os"],["œuf","œuf","œuf"],["jeu","jeu","jeu"]],"norwegian-bokmal":[["user","usere","us"],["person","personer","person"],["box","boxer","box"],["city","cityer","city"],["mouse","mouser","mouse"],["child","childer","child"],["analysis","analysiser","analysis"],["quiz","quizer","quiz"],["ox","oxer","ox"],["bus","buser","bus"],["day","dayer","day"],["lady","ladyer","lady"],["knife","knifer","knife"],["wolf","wolfer","wolf"],["man","maner","man"],["woman","womaner","woman"],["tooth","toother","tooth"],["foot","footer","foot"],["cactus","cactuser","cactus"],["criterion","criterioner","criterion"],["matrix","matrixer","matrix"],["vertex","vertexer","vertex"],["cat","cater","cat"],["dog","doger","dog"],["house","houser","house"],["cow","cower","cow"],["potato","potatoer","potato"],["hero","heroer","hero"],["status","statuser","status"],["news","newser","news"],["animal","animaler","animal"],["mal","maler","mal"],["cidadão","cidadãoer","cidadão"],["pão","pãoer","pão"],["país","paíser","país"],["mês","mêser","mês"],["lápis","lápiser","lápis"],["cavalo","cavaloer","cavalo"],["flor","flore","flor"],["papel","papeler","papel"],["canal","canaler","canal"],["luz","luzer","luz"],["homem","homemer","homem"],["alemão","alemãoer","alemão"],["livro","livroer","livro"],["ônibus","ônibuser","ônibus"],["cantor","cantore","cantor"],["amigo","amigoer","amigo"],["amiga","amigaer","amiga"],["giorno","giornoer","giorno"],["uomo","uomoer","uomo"],["città","cittàer","città"],["libro","libroer","libro"],["lago","lagoer","lago"],["amico","amicoer","amico"],["dio","dioer","dio"],["gatto","gattoer","gatto"],["bambino","bambinoer","bambino"],["casa","casaer","casa"],["hundo","hundoer","hundo"],["katt","katter","katt"],["bok","boker","bok"],["mann","manner","mann"],["dag","dager","dag"],["bil","biler","bil"],["gutt","gutter","gutt"],["barn","barn","barn"],["arm","armer","arm"],["boek","boeker","boek"],["estudiante","estudianter","estudiante"],["lápiz","lápizer","lápiz"],["corazón","corazóner","corazón"],["ciudad","ciudader","ciudad"],["mes","meser","mes"],["rey","reyer","rey"],["mujer","mujere","muj"],["canción","cancióner","canción"],["ej","ejer","ej"],["el","eler","el"],["elma","elmaer","elma"],["kitap","kitaper","kitap"],["göz","gözer","göz"],["ev","ever","ev"],["çocuk","çocuker","çocuk"],["adam","adamer","adam"],["kadın","kadıner","kadın"],["araba","arabaer","araba"],["okul","okuler","okul"],["monato","monatoer","monato"],["knabo","knaboer","knabo"],["hundoj","hundojer","hundoj"],["katoj","katojer","katoj"],["chat","chater","chat"],["cheval","chevaler","cheval"],["journal","journaler","journal"],["animaux","animauxer","animaux"],["chou","chouer","chou"],["bijou","bijouer","bijou"],["travail","travailer","travail"],["oeil","oeiler","oeil"],["nez","nezer","nez"],["pays","payser","pays"],["roi","roier","roi"],["prix","prixer","prix"],["feu","feuer","feu"],["lieu","lieuer","lieu"],["os","oser","os"],["œuf","œufer","œuf"],["jeu","jeuer","jeu"]],"portuguese":[["user","useres","user"],["person","persons","person"],["box","boxs","box"],["city","citys","city"],["mouse","mouses","mouse"],["child","childs","child"],["analysis","analysis","analysi"],["quiz","quizes","quiz"],["ox","oxs","ox"],["bus","bus","bu"],["day","days","day"],["lady","ladys","lady"],["knife","knifes","knife"],["wolf","wolfs","wolf"],["man","mans","man"],["woman","womans","woman"],["tooth","tooths","tooth"],["foot","foots","foot"],["cactus","cactus","cactu"],["criterion","criterions","criterion"],["matrix","matrixs","matrix"],["vertex","vertexs","vertex"],["cat","cats","cat"],["dog","dogs","dog"],["house","houses","house"],["cow","cows","cow"],["potato","potatos","potato"],["hero","heros","hero"],["status","status","statu"],["news","news","new"],["animal","animais","animal"],["mal","males","mal"],["cidadão","cidadãos","cidadão"],["pão","pães","pão"],["país","países","paí"],["mês","meses","mês"],["lápis","lápis","lápis"],["cavalo","cavalos","cavalo"],["flor","flores","flor"],["papel","papeis","papel"],["canal","canais","canal"],["luz","luzes","luz"],["homem","homens","homem"],["alemão","alemães","alemão"],["livro","livros","livro"],["ônibus","ônibus","ônibus"],["cantor","cantores","cantor"],["amigo","amigos","amigo"],["amiga","amigas","amiga"],["giorno","giornos","giorno"],["uomo","uomos","uomo"],["città","cittàs","città"],["libro","libros","libro"],["lago","lagos","lago"],["amico","amicos","amico"],["dio","dios","dio"],["gatto","gattos","gatto"],["bambino","bambinos","bambino"],["casa","casas","casa"],["hundo","hundos","hundo"],["katt","katts","katt"],["bok","boks","bok"],["mann","manns","mann"],["dag","dags","dag"],["bil","bis","bil"],["gutt","gutts","gutt"],["barn","barns","barn"],["arm","arns","arm"],["boek","boeks","boek"],["estudiante","estudiantes","estudiante"],["lápiz","lápizes","lápiz"],["corazón","corazóns","corazón"],["ciudad","ciudads","ciudad"],["mes","mes","me"],["rey","reys","rey"],["mujer","mujeres","mujer"],["canción","cancións","canción"],["ej","ejs","ej"],["el","eis","el"],["elma","elmas","elma"],["kitap","kitaps","kitap"],["göz","gözes","göz"],["ev","evs","ev"],["çocuk","çocuks","çocuk"],["adam","adans","adam"],["kadın","kadıns","kadın"],["araba","arabas","araba"],["okul","okuis","okul"],["monato","monatos","monato"],["knabo","knabos","knabo"],["hundoj","hundojs","hundoj"],["katoj","katojs","katoj"],["chat","chats","chat"],["cheval","chevais","cheval"],["journal","journais","journal"],["animaux","animauxs","animaux"],["chou","chous","chou"],["bijou","bijous","bijou"],["travail","travails","travail"],["oeil","oeils","oeil"],["nez","nezes","nez"],["pays","pays","pay"],["roi","rois","roi"],["prix","prixs","prix"],["feu","feus","feu"],["lieu","lieus","lieu"],["os","os","o"],["œuf","œufs","œuf"],["jeu","jeus","jeu"]],"spanish":[["user","useres","user"],["person","persones","person"],["box","boxes","box"],["city","cityes","city"],["mouse","mouses","mouse"],["child","childes","child"],["analysis","analysis","analysi"],["quiz","quices","quiz"],["ox","oxes","ox"],["bus","bus","bu"],["day","dayes","day"],["lady","ladyes","lady"],["knife","knifes","knife"],["wolf","wolfes","wolf"],["man","manes","man"],["woman","womanes","woman"],["tooth","toothes","tooth"],["foot","footes","foot"],["cactus","cactus","cactu"],["criterion","criteriones","criterion"],["matrix","matrixes","matrix"],["vertex","vertexes","vertex"],["cat","cates","cat"],["dog","doges","dog"],["house","houses","house"],["cow","cowes","cow"],["potato","potatos","potato"],["hero","heros","hero"],["status","status","statu"],["news","newses","new"],["animal","animales","animal"],["mal","males","mal"],["cidadão","cidadãos","cidadão"],["pão","pãos","pão"],["país","paises","paí"],["mês","mêses","mê"],["lápis","lápis","lápi"],["cavalo","cavalos","cavalo"],["flor","flores","flor"],["papel","papeles","papel"],["canal","canales","canal"],["luz","luces","luz"],["homem","homemes","homem"],["alemão","alemãos","alemão"],["livro","livros","livro"],["ônibus","ônibus","ônibu"],["cantor","cantores","cantor"],["amigo","amigos","amigo"],["amiga","amigas","amiga"],["giorno","giornos","giorno"],["uomo","uomos","uomo"],["città","cittàes","città"],["libro","libros","libro"],["lago","lagos","lago"],["amico","amicos","amico"],["dio","dios","dio"],["gatto","gattos","gatto"],["bambino","bambinos","bambino"],["casa","casas","casa"],["hundo","hundos","hundo"],["katt","kattes","katt"],["bok","bokes","bok"],["mann","mannes","mann"],["dag","dages","dag"],["bil","biles","bil"],["gutt","guttes","gutt"],["barn","barnes","barn"],["arm","armes","arm"],["boek","boekes","boek"],["estudiante","estudiantes","estudiante"],["lápiz","lápices","lápiz"],["corazón","corazones","corazón"],["ciudad","ciudades","ciudad"],["mes","meses","m"],["rey","reyes","rey"],["mujer","mujeres","mujer"],["canción","canciones","canción"],["ej","ejes","ej"],["el","los","el"],["elma","elmas","elma"],["kitap","kitapes","kitap"],["göz","göces","göz"],["ev","eves","ev"],["çocuk","çocukes","çocuk"],["adam","adames","adam"],["kadın","kadınes","kadın"],["araba","arabas","araba"],["okul","okules","okul"],["monato","monatos","monato"],["knabo","knabos","knabo"],["hundoj","hundojes","hundoj"],["katoj","katojes","katoj"],["chat","chates","chat"],["cheval","chevales","cheval"],["journal","journales","journal"],["animaux","animauxes","animaux"],["chou","chous","chou"],["bijou","bijous","bijou"],["travail","travailes","travail"],["oeil","oeiles","oeil"],["nez","neces","nez"],["pays","payses","pay"],["roi","rois","roi"],["prix","prixes","prix"],["feu","feus","feu"],["lieu","lieus","lieu"],["os","os","o"],["œuf","œufes","œuf"],["jeu","jeus","jeu"]],"turkish":[["user","userler","user"],["person","personlar","person"],["box","boxlar","box"],["city","cityler","city"],["mouse","mouseler","mouse"],["child","childler","child"],["analysis","analysisler","analysis"],["quiz","quizler","quiz"],["ox","oxlar","ox"],["bus","buslar","bus"],["day","daylar","day"],["lady","ladylar","lady"],["knife","knifeler","knife"],["wolf","wolflar","wolf"],["man","manlar","man"],["woman","womanlar","woman"],["tooth","toothlar","tooth"],["foot","footlar","foot"],["cactus","cactuslar","cactus"],["criterion","criterionlar","criterion"],["matrix","matrixler","matrix"],["vertex","vertexler","vertex"],["cat","catlar","cat"],["dog","doglar","dog"],["house","houseler","house"],["cow","cowlar","cow"],["potato","potatolar","potato"],["hero","herolar","hero"],["status","statuslar","status"],["news","newsler","news"],["animal","animallar","animal"],["mal","mallar","mal"],["cidadão","cidadãolar","cidadão"],["pão","pãolar","pão"],["país","paíslar","país"],["mês","mês","mês"],["lápis","lápisler","lápis"],["cavalo","cavalolar","cavalo"],["flor","florlar","flor"],["papel","papeller","papel"],["canal","canallar","canal"],["luz","luzlar","luz"],["homem","homemler","homem"],["alemão","alemãolar","alemão"],["livro","livrolar","livro"],["ônibus","ônibuslar","ônibus"],["cantor","cantorlar","cantor"],["amigo","amigolar","amigo"],["amiga","amigalar","amiga"],["giorno","giornolar","giorno"],["uomo","uomolar","uomo"],["città","cittàler","città"],["libro","librolar","libro"],["lago","lagolar","lago"],["amico","amicolar","amico"],["dio","diolar","dio"],["gatto","gattolar","gatto"],["bambino","bambinolar","bambino"],["casa","casalar","casa"],["hundo","hundolar","hundo"],["katt","kattlar","katt"],["bok","boklar","bok"],["mann","mannlar","mann"],["dag","daglar","dag"],["bil","biller","bil"],["gutt","guttlar","gutt"],["barn","barnlar","barn"],["arm","armlar","arm"],["boek","boekler","boek"],["estudiante","estudianteler","estudiante"],["lápiz","lápizler","lápiz"],["corazón","corazónlar","corazón"],["ciudad","ciudadlar","ciudad"],["mes","mesler","mes"],["rey","reyler","rey"],["mujer","mujerler","mujer"],["canción","canciónler","canción"],["ej","ejler","ej"],["el","eller","el"],["elma","elmalar","elma"],["kitap","kitaplar","kitap"],["göz","gözler","göz"],["ev","evler","ev"],["çocuk","çocuklar","çocuk"],["adam","adamlar","adam"],["kadın","kadınlar","kadın"],["araba","arabalar","araba"],["okul","okullar","okul"],["monato","monatolar","monato"],["knabo","knabolar","knabo"],["hundoj","hundojlar","hundoj"],["katoj","katojlar","katoj"],["chat","chatlar","chat"],["cheval","chevallar","cheval"],["journal","journallar","journal"],["animaux","animauxlar","animaux"],["chou","choular","chou"],["bijou","bijoular","bijou"],["travail","travailler","travail"],["oeil","oeiller","oeil"],["nez","nezler","nez"],["pays","payslar","pays"],["roi","roiler","roi"],["prix","prixler","prix"],["feu","feular","feu"],["lieu","lieular","lieu"],["os","oslar","os"],["œuf","œuflar","œuf"],["jeu","jeular","jeu"]]}
JSON;

    public static function languages(): array
    {
        $o = [];
        foreach (array_keys(json_decode(self::DATA, true)) as $l) { $o[$l] = [$l]; }
        return $o;
    }
    #[DataProvider('languages')]
    public function testLanguageWords(string $lang): void
    {
        $rows = json_decode(self::DATA, true)[$lang];
        $i = InflectorFactory::createForLanguage($lang)->build();
        foreach ($rows as [$w, $p, $s]) {
            $this->assertSame($p, $i->pluralize($w), "$lang plural $w");
            $this->assertSame($s, $i->singularize($w), "$lang singular $w");
            // cached second call
            $this->assertSame($p, $i->pluralize($w));
        }
    }
    public function testUnsupportedLanguage(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        $this->expectExceptionMessage('Language "klingon" is not supported.');
        InflectorFactory::createForLanguage('klingon');
    }
    public function testLanguageConstants(): void
    {
        foreach ([Language::ENGLISH, Language::ESPERANTO, Language::FRENCH, Language::ITALIAN, Language::NORWEGIAN_BOKMAL, Language::PORTUGUESE, Language::SPANISH, Language::TURKISH] as $l) {
            $this->assertInstanceOf(Inflector::class, InflectorFactory::createForLanguage($l)->build());
        }
        $this->assertInstanceOf(Inflector::class, InflectorFactory::create()->build());
    }
    public static function rulesNamespaces(): array
    {
        $o = [];
        foreach (['English', 'Esperanto', 'French', 'Italian', 'NorwegianBokmal', 'Portuguese', 'Spanish', 'Turkish'] as $n) { $o[$n] = [$n]; }
        return $o;
    }
    #[DataProvider('rulesNamespaces')]
    public function testRulesAndInflectibles(string $n): void
    {
        $rules = "Doctrine\\Inflector\\Rules\\$n\\Rules";
        $infl = "Doctrine\\Inflector\\Rules\\$n\\Inflectible";
        $unin = "Doctrine\\Inflector\\Rules\\$n\\Uninflected";
        $this->assertInstanceOf(Ruleset::class, $rules::getSingularRuleset());
        $this->assertInstanceOf(Ruleset::class, $rules::getPluralRuleset());
        foreach ($infl::getSingular() as $t) { $this->assertInstanceOf(Transformation::class, $t); }
        foreach ($infl::getPlural() as $t) { $this->assertInstanceOf(Transformation::class, $t); }
        foreach ($infl::getIrregular() as $t) { $this->assertInstanceOf(Substitution::class, $t); }
        foreach ($unin::getSingular() as $t) { $this->assertInstanceOf(Pattern::class, $t); }
        foreach ($unin::getPlural() as $t) { $this->assertInstanceOf(Pattern::class, $t); }
        $f = "Doctrine\\Inflector\\Rules\\$n\\InflectorFactory";
        $this->assertInstanceOf(\Doctrine\Inflector\LanguageInflectorFactory::class, new $f());
    }
    public function testNoopAndCached(): void
    {
        $n = new NoopWordInflector();
        $this->assertSame('abc', $n->inflect('abc'));
        $calls = 0;
        $inner = new class ($calls) implements WordInflector {
            public function __construct(public int &$calls) {}
            public function inflect(string $word): string { $this->calls++; return strtoupper($word); }
        };
        $c = new CachedWordInflector($inner);
        $this->assertSame('ABC', $c->inflect('abc'));
        $this->assertSame('ABC', $c->inflect('abc'));
        $this->assertSame(1, $calls);
    }
    public function testFactoryCustomRules(): void
    {
        $rs = new Ruleset(
            new Transformations(new Transformation(new Pattern('(.*)x$'), '\1xes')),
            new Patterns(new Pattern('fish')),
            new Substitutions(new Substitution(new Word('foo'), new Word('bar')))
        );
        $f = InflectorFactory::create()->withPluralRules($rs);
        $i = $f->build();
        $this->assertSame('bar', $i->pluralize('foo'));
        $this->assertSame('fish', $i->pluralize('fish'));
        $this->assertSame('users', $i->pluralize('user'));
        $f2 = InflectorFactory::create()->withPluralRules($rs, true)->withSingularRules($rs, true);
        $i2 = $f2->build();
        $this->assertSame('users', $i2->pluralize('users'));
        $i3 = InflectorFactory::create()->withSingularRules(null)->withPluralRules(null)->build();
        $this->assertSame('cat', $i3->singularize('cat'));
        $this->assertSame('cats', $i3->pluralize('cat'));
    }
    public function testRulesetInflectorOrder(): void
    {
        $a = new Ruleset(new Transformations(new Transformation(new Pattern('a$'), 'A')), new Patterns(), new Substitutions());
        $b = new Ruleset(new Transformations(new Transformation(new Pattern('a$'), 'B')), new Patterns(), new Substitutions());
        $this->assertSame('xA', (new RulesetInflector($a, $b))->inflect('xa'));
        $this->assertSame('xB', (new RulesetInflector($b, $a))->inflect('xa'));
        $this->assertSame('unmatched', (new RulesetInflector($a))->inflect('unmatched'));
    }
    public function testSubstitutionsFlippedAndCase(): void
    {
        $s = new Substitutions(new Substitution(new Word('man'), new Word('men')), new Substitution(new Word('goose'), new Word('geese')));
        $i = new RulesetInflector(new Ruleset(new Transformations(), new Patterns(), $s));
        $this->assertSame('men', $i->inflect('man'));
        $this->assertSame('Men', $i->inflect('Man'));
        $this->assertSame('Geese', $i->inflect('Goose'));
        $this->assertSame('other', $i->inflect('other'));
        $this->assertSame('', $i->inflect(''));
        $fl = new RulesetInflector(new Ruleset(new Transformations(), new Patterns(), $s->getFlippedSubstitutions()));
        $this->assertSame('man', $fl->inflect('men'));
        $this->assertSame('Goose', $fl->inflect('Geese'));
    }
    public function testPatternsMatchAndWord(): void
    {
        $p = new Patterns(new Pattern('sheep'), new Pattern('.*fish'));
        $this->assertTrue($p->matches('sheep'));
        $this->assertTrue($p->matches('goldfish'));
        $this->assertFalse($p->matches('cow'));
        $this->assertFalse((new Patterns())->matches('x'));
        $this->assertSame('foo', (new Word('foo'))->getWord());
        $pat = new Pattern('a.c');
        $this->assertTrue($pat->matches('abc'));
        $this->assertTrue($pat->matches('ABC'));
        $this->assertFalse($pat->matches('xyz'));
        $this->assertSame('a.c', $pat->getPattern());
        $this->assertStringContainsString('a.c', $pat->getRegex());
        $t = new Transformation(new Pattern('(.+)y$'), '\1ies');
        $this->assertSame('cities', $t->inflect('city'));
        $this->assertSame('xxx', $t->inflect('xxx'));
        $this->assertSame($t->getPattern()->getPattern(), '(.+)y$');
        $this->assertSame('\1ies', $t->getReplacement());
        $ss = new Substitution(new Word('a'), new Word('b'));
        $this->assertSame('a', $ss->getFrom()->getWord());
        $this->assertSame('b', $ss->getTo()->getWord());
        $this->assertSame('cities', (new Transformations($t))->inflect('city'));
        $this->assertSame('plain', (new Transformations($t))->inflect('plain'));
        $this->assertSame('x', (new Transformations())->inflect('x'));
    }
    public function testUnaccentVariants(): void
    {
        $i = InflectorFactory::create()->build();
        $this->assertSame('aeoeuess', $i->unaccent('äöüß'));
        $this->assertSame('Cafe', $i->unaccent('Café'));
        $this->assertSame('plain', $i->unaccent('plain'));
        $this->assertSame('Cafe', $i->unaccent("Caf\xE9"));
        $this->assertSame('AEOETHss', $i->unaccent("\xC6\x8C\xDE\xDF"));
        $this->assertFalse($i->seemsUtf8("\xE9"));
        $this->assertFalse($i->seemsUtf8("\xC3"));
        $this->assertFalse($i->seemsUtf8("\xFF"));
        $this->assertTrue($i->seemsUtf8("\xE2\x82\xAC"));
        $this->assertTrue($i->seemsUtf8("\xF0\x9F\x98\x80"));
        $this->assertTrue($i->seemsUtf8("\xF8\x88\x80\x80\x80"));
        $this->assertTrue($i->seemsUtf8("\xFC\x84\x80\x80\x80\x80"));
        $this->assertTrue($i->seemsUtf8('ascii'));
    }
    public function testCaseHelpers(): void
    {
        $i = InflectorFactory::create()->build();
        $this->assertSame('foo_bar_baz', $i->tableize('FooBarBaz'));
        $this->assertSame('FooBar', $i->classify('foo_bar'));
        $this->assertSame('fooBar', $i->camelize('foo-bar'));
        $this->assertSame('Foo Bar', $i->capitalize('foo bar'));
        $this->assertSame('Foo-Bar', $i->capitalize('foo-bar'));
        $this->assertSame('Foo-bar', $i->capitalize('foo-bar', ' '));
        $this->assertSame('hello-world', $i->urlize('Hello World'));
        $this->assertSame('cafe-au-lait', $i->urlize('Café au Lait!'));
        $this->assertSame('foobarbaz', $i->urlize('fooBarBaz'));
        $this->assertSame('xmlhttprequest', $i->urlize('XMLHttpRequest'));
        $this->assertSame('a-b-c', $i->urlize('a/b c'));
    }
}
