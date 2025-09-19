namespace GameLogic
{
    public interface ICodec
    {
        byte[] Encode(object message);
        object Decode(byte[] data);
    }
}
