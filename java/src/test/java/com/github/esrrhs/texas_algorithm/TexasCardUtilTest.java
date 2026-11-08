package com.github.esrrhs.texas_algorithm;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.Arrays;
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

        List<Byte> pokes = TexasAlgorithmUtil.strToPokes("方2,梅3,黑2,黑4,黑5");
        assertEquals(5, pokes.size());
        assertEquals("方2梅3黑2黑4黑5", TexasAlgorithmUtil.pokesToStr(pokes));
    }

    @Test
    @DisplayName("Test that the removed joker is rejected everywhere")
    public void testJokerRejected()
    {
        // 字符串层面的 "鬼" 已不属于 52 张牌堆
        assertThrows(IllegalArgumentException.class, () -> TexasAlgorithmUtil.strToPoke("鬼"));
        assertThrows(IllegalArgumentException.class, () -> TexasAlgorithmUtil.strToPokes("方2,梅3,黑2,黑4,鬼"));
        assertThrows(IllegalArgumentException.class, () -> TexasAlgorithmUtil.strToPokes("方2,梅3,黑2,黑4,黑5,鬼"));
        assertThrows(IllegalArgumentException.class, () -> TexasAlgorithmUtil.strToPokes("方2,梅3,黑2,黑4,黑5,黑6,鬼"));
        assertThrows(IllegalArgumentException.class, () -> TexasAlgorithmUtil.getMax("方2,梅3,黑2,黑4,鬼"));

        // 旧版鬼牌字节 0x58 = 88（花色 5）非法
        byte[] jokerBytes = { (byte) 0x12, (byte) 0x13, (byte) 0x24, (byte) 0x25, (byte) 0x36, (byte) 0x37,
                (byte) 0x58 };
        assertThrows(IllegalArgumentException.class, () -> TexasAlgorithmUtil.checkPokes(jokerBytes));

        List<Byte> jokerPokes = new ArrayList<>();
        for (byte value : jokerBytes)
        {
            jokerPokes.add(value);
        }
        assertThrows(IllegalArgumentException.class, () -> TexasAlgorithmUtil.getMax(jokerPokes));
        assertThrows(IllegalArgumentException.class, () -> TexasAlgorithmUtil.getWinPosition(jokerPokes));
        assertThrows(IllegalArgumentException.class, () -> TexasAlgorithmUtil.getWinType(jokerPokes));

        List<Byte> handWithJoker = Arrays.asList((byte) 0x12, (byte) 0x58);
        List<Byte> pub = Arrays.asList((byte) 0x13, (byte) 0x24, (byte) 0x25, (byte) 0x36);
        assertThrows(IllegalArgumentException.class, () -> TexasAlgorithmUtil.getMax(handWithJoker, pub));

        List<Byte> legalPokes = Arrays.asList((byte) 0x12, (byte) 0x13, (byte) 0x24, (byte) 0x25, (byte) 0x36,
                (byte) 0x37, (byte) 0x48);
        assertThrows(IllegalArgumentException.class, () -> TexasAlgorithmUtil.compare(jokerPokes, legalPokes));

        // key 中含鬼牌同样被拒绝
        long jokerKey = GenUtil.genCardBind(jokerPokes);
        assertThrows(IllegalArgumentException.class, () -> TexasAlgorithmUtil.getKeyData(jokerKey));
        assertThrows(IllegalArgumentException.class,
                () -> TexasAlgorithmUtil.compare(jokerKey, GenUtil.genCardBind(legalPokes)));
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
