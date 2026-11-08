package com.github.esrrhs.texas_algorithm;

import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.io.File;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.ConcurrentHashMap;

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

        File winRateFile = new File("texas_data_winrate_0.txt");
        if (winRateFile.exists())
        {
            TexasAlgorithmUtil.loadProbility();
            probDataAvailable = TexasAlgorithmUtil.isProbabilityLoaded();
        }
    }

    @Test
    @DisplayName("Test string and byte conversion utilities")
    public void testConversions()
    {
        String original = "方A,黑K,红10,梅J,黑9";
        List<Byte> pokes = TexasAlgorithmUtil.strToPokes(original);
        assertEquals(5, pokes.size());

        String reconstructed = TexasAlgorithmUtil.pokesToStr(pokes);
        assertEquals("方A黑K红10梅J黑9", reconstructed);

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
        String best = TexasAlgorithmUtil.getMax("方4,方2", "黑2,黑A,方3,黑5,黑6");
        assertNotNull(best);
        // Best 5 cards should form a straight: 2, 3, 4, 5, 6
        assertTrue(best.contains("2") && best.contains("3") && best.contains("4") && best.contains("5") && best.contains("6"));
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
        assertTrue(p1 > 0.55f && p1 < 0.56f, "p1 should be around 0.5533 (exact), got: " + p1);

        float p2 = TexasAlgorithmUtil.getHandProbability("方2,方3", "");
        assertTrue(p2 > 0.35f && p2 < 0.37f, "p2 should be around 0.3598 (exact), got: " + p2);
    }

    @Test
    @DisplayName("Test graceful handling of unloaded probability data")
    public void testUnloadedProbabilityGraceful()
    {
        // A wrong number of hole cards must be reported as an exception, never a crash.
        try
        {
            TexasAlgorithmUtil.getHandProbability(Arrays.asList((byte) 0x02), Collections.emptyList());
            fail("getHandProbability with one hole card should throw");
        }
        catch (IllegalArgumentException expected)
        {
            // expected
        }
    }

    @Test
    @DisplayName("Missing win-rate tables throw instead of crashing")
    @SuppressWarnings("unchecked")
    public void testMissingWinRateTableGraceful()
    {
        if (!dataAvailable)
        {
            System.out.println("Skipping testMissingWinRateTableGraceful: data files not present");
            return;
        }

        ConcurrentHashMap<Long, TexasAlgorithmUtil.WinRateData>[] saved =
                (ConcurrentHashMap<Long, TexasAlgorithmUtil.WinRateData>[]) new ConcurrentHashMap[3];
        for (int i = 0; i < 3; i++)
        {
            saved[i] = TexasAlgorithmUtil.winRateMap[i];
            TexasAlgorithmUtil.winRateMap[i] = new ConcurrentHashMap<>();
        }

        try
        {
            try
            {
                TexasAlgorithmUtil.getHandProbability(Arrays.asList((byte) 0x02, (byte) 0x03),
                        Collections.emptyList());
                fail("preflop without winrate_0 should throw");
            }
            catch (IllegalArgumentException expected)
            {
                // expected
            }

            float p = TexasAlgorithmUtil.getHandProbability(Arrays.asList((byte) 0x02, (byte) 0x03),
                    Arrays.asList((byte) 0x04, (byte) 0x05, (byte) 0x06));
            assertTrue(p > 0 && p < 1, "flop should be enumerated without tables");
        }
        finally
        {
            for (int i = 0; i < 3; i++)
            {
                TexasAlgorithmUtil.winRateMap[i] = saved[i];
            }
        }
    }
}
