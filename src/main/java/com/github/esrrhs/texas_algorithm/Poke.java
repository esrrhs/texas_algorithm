package com.github.esrrhs.texas_algorithm;

public final class Poke implements Cloneable
{
	public static final byte PokeValue_2 = 2;
	public static final byte PokeValue_3 = 3;
	public static final byte PokeValue_4 = 4;
	public static final byte PokeValue_5 = 5;
	public static final byte PokeValue_6 = 6;
	public static final byte PokeValue_7 = 7;
	public static final byte PokeValue_8 = 8;
	public static final byte PokeValue_9 = 9;
	public static final byte PokeValue_10 = 10;
	public static final byte PokeValue_J = 11;
	public static final byte PokeValue_Q = 12;
	public static final byte PokeValue_K = 13;
	public static final byte PokeValue_A = 14;
	public static final byte[] PokeValues = new byte[]
	{ PokeValue_2, PokeValue_3, PokeValue_4, PokeValue_5, PokeValue_6, PokeValue_7, PokeValue_8, PokeValue_9,
			PokeValue_10, PokeValue_J, PokeValue_Q, PokeValue_K, PokeValue_A };

	public static final byte PokeColor_HEI = 3;
	public static final byte PokeColor_HONG = 2;
	public static final byte PokeColor_MEI = 1;
	public static final byte PokeColor_FANG = 0;

	public static final Poke POKE_INVALID = new Poke((byte) 0, (byte) 0);

	public byte color;
	public byte value;

	public Poke(byte color, byte value)
	{
		this.color = color;
		this.value = value;
	}

	public Poke(byte byteValue)
	{
		this.color = (byte) (byteValue >> 4);
		this.value = (byte) (byteValue % 16);
	}

	public void setColor(byte color)
	{
		this.color = color;
	}

	public void setValue(byte value)
	{
		this.value = value;
	}

	public byte getColor()
	{
		return this.color;
	}

	public byte getValue()
	{
		return this.value;
	}

	public byte toByte()
	{
		return (byte) (color << 4 | value);
	}

	@Override
	public boolean equals(Object object)
	{
		if (!(object instanceof Poke))
		{
			return false;
		}
		Poke other = (Poke) object;
		return this.color == other.color && this.value == other.value;
	}

	@Override
	public String toString()
	{
		return String.format("%d:%d", this.color, this.value);
	}

	@Override
	public Poke clone()
	{
		return new Poke(this.color, this.value);
	}
}
