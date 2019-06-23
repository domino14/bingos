This codebase is for the calculations of type I, type II and type III bingos in Scrabble. These terms are taken from Michael Baron's wordbook, and we use his methodology for the calculation of MMPR, a measure of the "power" of a stem.

His methodology results in slightly different results. Taking OWL2 as a lexicon, and comparing the generated stems, we can see there is a slight disagreement at number 5:

|#|alpha|MSP|UT|MMPR|
|-|-----|---|--|----|
|5|EINORS|1.3333|68|90.6667|

Whereas the book (https://www.amazon.com/SCRABBLE%C2%99-Wordbook-Mike-Baron/dp/1402750862) cites the following:

|#|alpha|MSP|UT|MMPR|
|-|-----|---|--|----|
|5|EINORS|1.333|68|90.644|

The 90.644 seems to have come from 1.333 * 68 (as opposed to using the full number and multiplying 4/3 * 68).

This should normally not make a difference, but it messes up some tiebreakers. For example, stems 98 and 99 (EIRSTU and AEILOS) have an MMPR of 33.350 and 33.325, relatively, according to the book, but in reality their MMPR is the exact same -- 33.3̅, so AEILOS should _actually_ come before EIRSTU because of its higher MSP (Modified Stem Probability; the actual normalized probability of the stem, modified to artificially make Ss 1.5 times more common).

For NWL18, the results are a bit different than for OWL2. Mike cites that Type III 7s should be made from the least probable stem in the top 100 + a letter that there are 2 of. In OWL2, that stem was TUNERS (MSP of 0.444) + H, for example, to make HUNTERS. Anything with a probability of HUNTERS that is not already a type I or type II 7 is a type III 7.

But in NWL18, the least probable stem has an MSP of 0.5; AEIPRS for example. Some previously lower stems were moved up by additions since OWL2 (for example, AEGORT did not use to appear in the top 100, but the addition of GOATIER and its 9 extra Usable Tiles made it leapfrog from 145th to 95th, displacing TUNERS in the top 100). Picking HARPIES as the cutoff does not seem good enough; there are around 1400 words in between HARPIES and HUNTERS by probability that one would completely miss. Therefore, I recommend that we keep HUNTERS as the arbitrary type III cutoff.

For type III 8s, the cutoff seems to be the probability of NOTIFIED. I'm actually not sure where this word comes from, maybe the cutoff should be analogous to that of type III 7s and we can pick a stem like AEGNRST (with its lowest MSP in the top 100 of 0.5). There are more than 2000 useful words in between ENGRAFTS and NOTIFIED; maybe the type III 8s should include these? As it stands, that list is pretty small right now.

Finally, type I 8s are defined as the words formed by the top 100 6-letter stems + 2 tiles, which is not exactly analogous to type I 7s, one might expect it to be the top 100 7-letter stems + 1 tile. But the former may be more useful, you get stuff like VILAYETS (SALTIE + VY) whereas that word wouldn't come up in a 7-letter stem.
