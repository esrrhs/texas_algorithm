package com.github.esrrhs.texas_algorithm;

import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.io.File;
import java.util.ArrayList;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

public class TexasAlgorithmUtilTest
{
    private static boolean dataAvailable = false;
    private static boolean probDataAvailable = false;

    @BeforeAll
    public static void setUp()
    {
        File colorFile = new File("texas_data_color.txt");
        if (colorFile.exists())
        {
            TexasAlgorithmUtil.load();
            dataAvailable = TexasAlgorithmUtil.isLoaded();
        }

        File optFile = new File("texas_data_opt_2.txt");
        if (optFile.exists())
        {
            TexasAlgorithmUtil.loadProbility();
            probDataAvailable = TexasAlgorithmUtil.isProbabilityLoaded();
        }
    }

    @Test
    @DisplayName("Test string and byte conversion utilities")
    public void testConversions()
    {
        String original = "方A,黑K,红10,梅J,鬼";
        List<Byte> pokes = TexasAlgorithmUtil.strToPokes(original);
        assertEquals(5, pokes.size());

        String reconstructed = TexasAlgorithmUtil.pokesToStr(pokes);
        assertEquals("方A黑K红10梅J鬼", reconstructed);

        long key = GenUtil.genCardBind(pokes);
        List<Byte> fromKey = TexasAlgorithmUtil.keyToPoke(key);
        assertEquals(5, fromKey.size());
    }

    @Test
    @DisplayName("Test getMax without wild cards")
    public void testGetMaxWithoutGui()
    {
        if (!dataAvailable)
        {
            System.out.println("Skipping testGetMaxWithoutGui: data files not present");
            return;
        }
        String best = TexasAlgorithmUtil.getMax("方4,方2", "黑2,黑A,方3,黑5,黑6", null);
        assertNotNull(best);
        // Best 5 cards should form a straight: 2, 3, 4, 5, 6
        assertTrue(best.contains("2") && best.contains("3") && best.contains("4") && best.contains("5") && best.contains("6"));
    }

    @Test
    @DisplayName("Test getMax with wild cards (Gui)")
    public void testGetMaxWithGui()
    {
        if (!dataAvailable)
        {
            System.out.println("Skipping testGetMaxWithGui: data files not present");
            return;
        }
        List<Byte> guiTrans = new ArrayList<>();
        String best = TexasAlgorithmUtil.getMax("方2,梅3,黑2,黑4,鬼", guiTrans);
        assertNotNull(best);
        // Wild card should turn this hand into three of a kind or full house
        assertFalse(guiTrans.isEmpty());

        guiTrans.clear();
        String best2 = TexasAlgorithmUtil.getMax("方2,方3,方4,鬼,鬼,黑6,红6", guiTrans);
        assertNotNull(best2);
        // Two wild cards with 2, 3, 4 of diamonds should transform into straight flush (5, 6 of diamonds)
        assertEquals(2, guiTrans.size());
    }

    @Test
    @DisplayName("Test table lookup evaluation when data files are available")
    public void testTableLookup()
    {
        if (!dataAvailable)
        {
            System.out.println("Skipping testTableLookup: data files not present");
            return;
        }

        String cards = "方4,方A,黑2,黑A,黑3,黑5,黑6";
        String cards1 = "红8,方A,方2,黑8,黑3,黑5,黑7";

        int pos = TexasAlgorithmUtil.getWinPosition(cards);
        assertTrue(pos > 0);
        assertEquals(4010, pos);

        double prob = TexasAlgorithmUtil.getWinProbability(cards);
        assertTrue(prob > 0.8 && prob < 1.0);

        int type = TexasAlgorithmUtil.getWinType(cards);
        assertEquals(TexasCardUtil.TEXAS_CARD_TYPE_TONGHUA, type);

        int pos1 = TexasAlgorithmUtil.getWinPosition(cards1);
        assertEquals(1143, pos1);
        int type1 = TexasAlgorithmUtil.getWinType(cards1);
        assertEquals(TexasCardUtil.TEXAS_CARD_TYPE_DUIZI, type1);

        int cmp = TexasAlgorithmUtil.compare(cards, cards1);
        assertTrue(cmp > 0, "cards should beat cards1");
    }

    @Test
    @DisplayName("Test win probability estimation when probability data is available")
    public void testProbabilityEstimation()
    {
        if (!probDataAvailable)
        {
            System.out.println("Skipping testProbabilityEstimation: prob data files not present");
            return;
        }

        float p1 = TexasAlgorithmUtil.getHandProbability("方3,方A", "黑2,黑4,黑5,黑K");
        assertTrue(p1 > 0.6f && p1 < 0.9f, "p1 should be around 0.74, got: " + p1);

        float p2 = TexasAlgorithmUtil.getHandProbability("方2,方3", "");
        assertTrue(p2 > 0.3f && p2 < 0.6f, "p2 should be around 0.45, got: " + p2);
    }

    @Test
    @DisplayName("Test graceful handling of unloaded probability data")
    public void testUnloadedProbabilityGraceful()
    {
        // When querying an invalid or unloaded card key
        TexasAlgorithmUtil.ProbilityData data = TexasAlgorithmUtil.getHandProbability(1L);
        assertNull(data);
    }
}
