package com.github.esrrhs.texas_algorithm;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

public class TexasCardUtilTest
{
    @Test
    @DisplayName("Test Poke encoding and string parsing")
    public void testPokeParsing()
    {
        byte b = TexasAlgorithmUtil.strToPoke("方A");
        Poke p = new Poke(b);
        assertEquals(Poke.PokeColor_FANG, p.getColor());
        assertEquals(Poke.PokeValue_A, p.getValue());
        assertEquals("方A", p.toString());

        byte guiByte = TexasAlgorithmUtil.strToPoke("鬼");
        Poke gui = new Poke(guiByte);
        assertTrue(gui.isGui());
        assertEquals(Poke.GUI.toByte(), gui.toByte());

        List<Byte> pokes = TexasAlgorithmUtil.strToPokes("方2,梅3,黑2,黑4,鬼");
        assertEquals(5, pokes.size());
        assertEquals("方2梅3黑2黑4鬼", TexasAlgorithmUtil.pokesToStr(pokes));
    }

    @Test
    @DisplayName("Test card types classification without lookup table")
    public void testCardTypes()
    {
        // 1. Royal Flush
        checkCardType("黑10,黑J,黑Q,黑K,黑A", TexasCardUtil.TEXAS_CARD_TYPE_KINGTONGHUASHUN);

        // 2. Straight Flush
        checkCardType("红9,红10,红J,红Q,红K", TexasCardUtil.TEXAS_CARD_TYPE_TONGHUASHUN);

        // 3. Four of a kind
        checkCardType("黑A,红A,梅A,方A,黑K", TexasCardUtil.TEXAS_CARD_TYPE_SITIAO);

        // 4. Full House
        checkCardType("黑A,红A,梅A,黑K,红K", TexasCardUtil.TEXAS_CARD_TYPE_HULU);

        // 5. Flush
        checkCardType("黑2,黑4,黑6,黑8,黑K", TexasCardUtil.TEXAS_CARD_TYPE_TONGHUA);

        // 6. Straight
        checkCardType("黑10,红J,梅Q,方K,黑A", TexasCardUtil.TEXAS_CARD_TYPE_SHUNZI);
        checkCardType("黑A,红2,梅3,方4,黑5", TexasCardUtil.TEXAS_CARD_TYPE_SHUNZI); // A-2-3-4-5

        // 7. Three of a kind
        checkCardType("黑A,红A,梅A,黑K,红Q", TexasCardUtil.TEXAS_CARD_TYPE_SANTIAO);

        // 8. Two Pair
        checkCardType("黑A,红A,黑K,红K,黑Q", TexasCardUtil.TEXAS_CARD_TYPE_LIANGDUI);

        // 9. One Pair
        checkCardType("黑A,红A,黑K,红Q,黑J", TexasCardUtil.TEXAS_CARD_TYPE_DUIZI);

        // 10. High Card
        checkCardType("黑A,红K,黑Q,红J,黑9", TexasCardUtil.TEXAS_CARD_TYPE_GAOPAI);
    }

    @Test
    @DisplayName("Test wild card (Gui) substitution in 5-card evaluation")
    public void testWildCardEvaluation()
    {
        // 4 cards + 1 Gui -> Royal Flush
        checkCardTypeWithGui("黑10,黑J,黑Q,黑K,鬼", TexasCardUtil.TEXAS_CARD_TYPE_KINGTONGHUASHUN);

        // 3 cards + 2 Gui -> Straight Flush
        checkCardTypeWithGui("方2,方3,方4,鬼,鬼", TexasCardUtil.TEXAS_CARD_TYPE_TONGHUASHUN);

        // Three of a kind + 1 Gui -> Four of a kind
        checkCardTypeWithGui("黑A,红A,梅A,黑K,鬼", TexasCardUtil.TEXAS_CARD_TYPE_SITIAO);

        // One pair + 1 Gui -> Three of a kind
        checkCardTypeWithGui("黑A,红A,黑K,红Q,鬼", TexasCardUtil.TEXAS_CARD_TYPE_SANTIAO);
    }

    @Test
    @DisplayName("Test hand comparison")
    public void testCardComparison()
    {
        ArrayList<Poke> royalFlush = toSortedPokeList("黑10,黑J,黑Q,黑K,黑A");
        ArrayList<Poke> straightFlush = toSortedPokeList("红9,红10,红J,红Q,红K");
        ArrayList<Poke> fourOfAKind = toSortedPokeList("黑A,红A,梅A,方A,黑K");
        ArrayList<Poke> fullHouse = toSortedPokeList("黑A,红A,梅A,黑K,红K");

        assertTrue(TexasCardUtil.compareCardsWithoutGui(royalFlush, straightFlush) > 0);
        assertTrue(TexasCardUtil.compareCardsWithoutGui(straightFlush, fourOfAKind) > 0);
        assertTrue(TexasCardUtil.compareCardsWithoutGui(fourOfAKind, fullHouse) > 0);
        assertEquals(0, TexasCardUtil.compareCardsWithoutGui(royalFlush, royalFlush));
    }

    private void checkCardType(String cardsStr, int expectedType)
    {
        ArrayList<Poke> pokes = toPokeList(cardsStr);
        int actualType = TexasCardUtil.getCardTypeUnorderedWithoutGui(pokes);
        assertEquals(expectedType, actualType, "Failed for cards: " + cardsStr);
    }

    private void checkCardTypeWithGui(String cardsStr, int expectedType)
    {
        ArrayList<Poke> pokes = toPokeList(cardsStr);
        ArrayList<Poke> picked = new ArrayList<>();
        TexasCardUtil.fiveFromFive(pokes, picked);
        int actualType = TexasCardUtil.getCardTypeUnorderedWithoutGui(picked);
        assertEquals(expectedType, actualType, "Failed with gui for cards: " + cardsStr);
    }

    private ArrayList<Poke> toPokeList(String cardsStr)
    {
        List<Byte> bytes = TexasAlgorithmUtil.strToPokes(cardsStr);
        return GenUtil.toArray(GenUtil.genCardBind(bytes));
    }

    private ArrayList<Poke> toSortedPokeList(String cardsStr)
    {
        ArrayList<Poke> list = toPokeList(cardsStr);
        ArrayList<Poke> picked = new ArrayList<>();
        TexasCardUtil.fiveFromFiveWithoutGui(list, picked);
        return picked;
    }
}
